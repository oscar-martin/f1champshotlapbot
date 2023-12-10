package livemap

import (
	"bufio"
	"encoding/json"
	"f1champshotlapsbot/pkg/layout"
	"f1champshotlapsbot/pkg/model"
	"f1champshotlapsbot/pkg/pubsub"
	"f1champshotlapsbot/pkg/resources"
	"fmt"
	"html/template"
	"image"
	"log"
	"math"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/llgcode/draw2d"
	"github.com/llgcode/draw2d/draw2dsvg"
)

type Point struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	Type int     `json:"type"`
}

type AIW []Point

var upgrader = websocket.Upgrader{} // use default options

type LiveMap struct {
	sessionRunning      bool
	serverId            string
	path                string
	svgTrackResource    resources.Resource
	selectedSessionData model.SelectedSessionData
	gc                  draw2d.GraphicContext
	svgMetadata         layout.SvgMetadata
	carsPositionChan    <-chan []model.CarPosition
	carsPosition        []model.CarPosition
	mu                  sync.Mutex
}

func NewLiveMap(r *mux.Router, serverId, path string) *LiveMap {
	lm := &LiveMap{
		serverId:         serverId,
		sessionRunning:   false,
		path:             path,
		carsPositionChan: pubsub.CarsPositionPubSub.Subscribe(pubsub.PubSubCarsPositionPreffix + serverId),
		carsPosition:     []model.CarPosition{},
		mu:               sync.Mutex{},
	}

	go lm.updateCarsPosition()

	lm.addHandlers(r, path)
	return lm
}

func (lm *LiveMap) GetPath() string {
	return lm.path
}

func (lm *LiveMap) updateCarsPosition() {
	for carsPosition := range lm.carsPositionChan {
		lm.mu.Lock()
		if !lm.sessionRunning {
			lm.mu.Unlock()
			continue
		}
		transformedCarsPosition := make([]model.CarPosition, len(carsPosition))
		for i, carPosition := range carsPosition {
			if lm.gc != nil {
				transformedCarsPosition[i] = carPosition
				p := lm.transformPosition(carPosition.X, carPosition.Z, layout.ScaleSVG)
				transformedCarsPosition[i].X = p.X
				transformedCarsPosition[i].Z = p.Z
			}
		}
		lm.carsPosition = transformedCarsPosition
		lm.mu.Unlock()
	}
}

func (lm *LiveMap) StartSession(ssd model.SelectedSessionData, svgTrackResource resources.Resource) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	lm.selectedSessionData = ssd
	lm.svgTrackResource = svgTrackResource
	svgPath := lm.svgTrackResource.FilePath()

	// open the svg file and read the three last lines
	svgFile, err := os.Open(svgPath)
	if err != nil {
		log.Printf("Error opening svg file: %s", err)
		return
	}
	defer svgFile.Close()

	// Create a scanner to read the file line by line
	scanner := bufio.NewScanner(svgFile)

	// Variables to store the last two lines
	var lastLine, secondLastLine string

	// Read the file line by line
	for scanner.Scan() {
		secondLastLine = lastLine
		lastLine = scanner.Text()
	}

	// Check for errors during scanning
	if err := scanner.Err(); err != nil {
		log.Println("Error reading file:", err)
		return
	}

	// Print the second-to-last line
	if secondLastLine != "" {
		err = json.Unmarshal([]byte(secondLastLine), &lm.svgMetadata)
		if err != nil {
			log.Printf("Error unmarshalling svg metadata: %s", err)
			return
		}
		lm.svgMetadata.Rect = image.Rect(0, 0, int(lm.svgMetadata.Width), int(lm.svgMetadata.Height))
		lm.gc = draw2dsvg.NewGraphicContext(draw2dsvg.NewSvg())
	} else {
		log.Println("The file does not have the required metadata")
	}

	log.Printf("LiveMap endpoint started for Server %s\n", lm.serverId)
	lm.sessionRunning = true
}

func (lm *LiveMap) StopSession() {
	lm.mu.Lock()
	defer lm.mu.Unlock()
	lm.sessionRunning = false
}

func (lm *LiveMap) websocketHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		defer c.Close()
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("recv: %s (%d)", message, mt)
		t := time.NewTicker(100 * time.Millisecond)
		for {
			select {
			case <-t.C:
				lm.mu.Lock()
				bytes, err := json.Marshal(lm.carsPosition)
				lm.mu.Unlock()
				if err != nil {
					log.Println("marshal:", err)
					return
				}
				err = c.WriteMessage(mt, bytes)
				if err != nil {
					log.Println("write:", err)
					return
				}
			case <-r.Context().Done():
				log.Print("websocket closed\n")
				t.Stop()
				return
			}
		}
	}
}

type Data struct {
	WebSocketURL string
	TrackURL     string
	Width        int
	Height       int
	Scale        float64
}

func (lm *LiveMap) livemapHandler(serverId string) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !lm.sessionRunning {
			fmt.Fprintf(w, "No hay sesiones activas")
			return
		} else if lm.svgTrackResource.IsZero() {
			fmt.Fprintf(w, "No se ha creado aún el mapa del trazado")
			return
		}
		e := Data{
			WebSocketURL: "ws://" + r.Host + serverId + "/livemap",
			TrackURL:     "http://" + r.Host + "/resources/" + lm.svgTrackResource.FileName(),
			Width:        int(lm.svgMetadata.Width),
			Height:       int(lm.svgMetadata.Height),
			Scale:        (1.0 - layout.ScaleSVG),
		}
		homeTemplate.Execute(w, e)
	}
}

func (lm *LiveMap) addHandlers(r *mux.Router, serverId string) {
	r.HandleFunc("/livemap", lm.websocketHandler())
	r.HandleFunc("/live", lm.livemapHandler(serverId))
}

// Flips the image around the Y axis.
func invertY(gc draw2d.GraphicContext, rect image.Rectangle, factor float64) {
	height := rect.Max.Y
	gc.Translate(0, float64(height))
	gc.Scale(1.0, -1.0)
}

func (lm *LiveMap) transformPosition(dataX, dataZ float64, factor float64) Point {
	lm.gc.Save()
	lm.gc.MoveTo(dataX*(1.0-factor)+lm.svgMetadata.OffsetX, dataZ*(1.0-factor)+lm.svgMetadata.OffsetZ)
	invertY(lm.gc, lm.svgMetadata.Rect, factor)
	if lm.svgMetadata.Rotate {
		lm.gc.Rotate(math.Pi / 2)
		f := lm.svgMetadata.Width / lm.svgMetadata.Height
		lm.gc.Translate(0, -f*float64(lm.svgMetadata.Rect.Max.Y))
	}
	x, y := lm.gc.LastPoint()
	lp := []float64{x, y}
	m := lm.gc.GetMatrixTransform()
	m.Transform(lp)
	lm.gc.Restore()
	return Point{X: lp[0], Y: 0.0, Z: lp[1], Type: 0}
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>rFactor2 LiveMap</title>
</head>
<body>

  <!-- SVG container -->
	<svg id="svgContainer" width="{{ .Width }}" height="{{ .Height }}" xmlns="http://www.w3.org/2000/svg"></svg>

  <script>
    const trackUrl = '{{ .TrackURL }}';
    const wsUrl = '{{ .WebSocketURL }}';

    // SVG container element
    const svgContainer = document.getElementById('svgContainer');

		const cars = new Map();

    // WebSocket connection
    const socket = new WebSocket(wsUrl);

    // Connection opened event
    socket.addEventListener('open', (event) => {
      console.log('WebSocket connection opened:', event);
			socket.send("start");
    });

    // Listen for messages from the server
    socket.addEventListener('message', (event) => {
      // Parse the received JSON data
      const driversData = JSON.parse(event.data);

			const driversAlive = new Set();

			var i = 0;
			for (const data of driversData) {
				const x = data.x;
				const y = data.z;
				const id = data.dri;

				// add driver as alive
				driversAlive.add(data.dri);

				var carElements = null;
				if (!cars.has(id)) {
					carElements = buildCar(id);
					cars.set(id, carElements);
				} else {
					carElements = cars.get(id);
				}
				if (i < driversData.length - 1) {
					drawCar(carElements, x, y, '#EEEEEE', '#393939');
				} else {
					// leader
					drawCar(carElements, x, y, '#E7E772', '#393939');
				}
				i++;
			}

			// finally, delete element from non-alive drivers
			allDrivers = cars.keys()
			for (const dri of allDrivers) {
				if (!driversAlive.has(dri)) {
					const carElements = cars.get(dri);
					carElements.car.remove();
					cars.delete(dri);
				}
			}
		});

    // Connection closed event
    socket.addEventListener('close', (event) => {
      console.log('WebSocket connection closed:', event);
    });

    // Connection error event
    socket.addEventListener('error', (event) => {
      console.error('WebSocket connection error:', event);
    });

		function buildCar(id) {
			const carElement = document.createElementNS('http://www.w3.org/2000/svg', 'g');
			const circleElement = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
			const textElement = document.createElementNS('http://www.w3.org/2000/svg', 'text');

			textElement.setAttribute('text-anchor', 'middle');
			textElement.setAttribute('dy', '.3em');
			textElement.setAttribute('stroke-width', '2px');
			textElement.setAttribute('font-size', '20px');
			textElement.textContent = id;
			circleElement.setAttribute('r', 25); // Radius, adjust as needed
			circleElement.setAttribute('stroke', '#111111'); // Color, adjust as needed
			circleElement.setAttribute('stroke-width', '2px');
			carElement.appendChild(circleElement);
			carElement.appendChild(textElement);

			svgContainer.appendChild(carElement);

			return {circle: circleElement, text: textElement, car: carElement};
		}

    // Function to draw a circle on the SVG
    function drawCar(carElements, x, y, bColor, fColor) {
			const circleElement = carElements.circle;
			const textElement = carElements.text;
			const carElement = carElements.car;

			textElement.setAttribute('x', x);
			textElement.setAttribute('y', y);
			textElement.setAttribute('stroke', fColor);
			circleElement.setAttribute('cx', x);
      circleElement.setAttribute('cy', y);
			circleElement.setAttribute('fill', bColor);
    }

    // Function to download and display the SVG
    async function downloadAndDisplaySVG(url) {
      try {
        // Fetch the SVG file
        const response = await fetch(url);

        if (!response.ok) {
          throw new Error(` + "`Failed to fetch SVG: ${response.statusText}`" + `);
        }

        // Get the SVG content as text
        const svgText = await response.text();

        // Insert the SVG into the container
        svgContainer.innerHTML = svgText;
      } catch (error) {
        console.error(error.message);
      }
    }

    // Call the function to download and display the SVG
    downloadAndDisplaySVG(trackUrl);
  </script>
</body>
</html>
`))

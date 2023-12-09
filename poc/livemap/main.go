package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/websocket"
)

type Point struct {
	X    float64 `json:"x"`
	Y    float64 `json:"y"`
	Z    float64 `json:"z"`
	Type int     `json:"type"`
}

type AIW []Point

var addr = flag.String("addr", ":8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(aiw AIW) func(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("recv: %s", message)
		t := time.NewTicker(50 * time.Millisecond)
		i := 0
		for _ = range t.C {
			if i >= len(aiw) {
				i = 0
			}
			bytes, err := json.Marshal(aiw[i])
			if err != nil {
				log.Println("marshal:", err)
				break
			}
			err = c.WriteMessage(mt, bytes)
			if err != nil {
				log.Println("write:", err)
				break
			}
			i++
		}
	}
}

type Data struct {
	WebSocketURL string
	TrackURL     string
	Width        int
	Height       int
}

func home(w http.ResponseWriter, r *http.Request) {
	e := Data{
		WebSocketURL: "ws://" + r.Host + "/cars",
		TrackURL:     "http://" + r.Host + "/track/imola.svg",
		Width:        2000,
		Height:       1000,
	}
	homeTemplate.Execute(w, e)
}

func main() {
	flag.Parse()
	log.SetFlags(0)

	carJsonFile, err := os.Open(fmt.Sprintf("./car.%s.json", "imola"))
	if err != nil {
		panic(err)
	}
	bytes, err := io.ReadAll(carJsonFile)
	if err != nil {
		panic(err)
	}
	var carAiw AIW
	err = json.Unmarshal(bytes, &carAiw)
	if err != nil {
		panic(err)
	}

	fs := http.FileServer(http.Dir("."))
	http.Handle("/track/", http.StripPrefix("/track/", fs))
	http.HandleFunc("/cars", echo(carAiw))
	http.HandleFunc("/", home)

	log.Fatal(http.ListenAndServe(*addr, nil))
}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>WebSocket SVG Drawing</title>
</head>
<body>

  <!-- SVG container -->
	<svg id="svgContainer" width="{{ .Width }}" height="{{ .Height }}" xmlns="http://www.w3.org/2000/svg"></svg>

  <script>
    // Replace with your server and WebSocket URLs
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
      const data = JSON.parse(event.data);

      // Access properties of the message
      const x = data.x;
      const y = data.z;

			const id1 = 32;
			var carElements = null;
			if (!cars.has(id1)) {
				carElements = buildCar(id1, '#ECE2D0', '#D3A588');
				cars.set(id1, carElements);
			} else {
				carElements = cars.get(id1);
			}
      drawCar(carElements, x, y);

			const id2 = 51;
			if (!cars.has(id2)) {
				carElements = buildCar(id2, '#A8B7AB', '#2F643A');
				cars.set(id2, carElements);
			} else {
				carElements = cars.get(id2);
			}
				drawCar(carElements, x+50, y+50);
		});

    // Connection closed event
    socket.addEventListener('close', (event) => {
      console.log('WebSocket connection closed:', event);
    });

    // Connection error event
    socket.addEventListener('error', (event) => {
      console.error('WebSocket connection error:', event);
    });

		function buildCar(id, bColor, fColor) {
			const carElement = document.createElementNS('http://www.w3.org/2000/svg', 'g');
			const circleElement = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
			const textElement = document.createElementNS('http://www.w3.org/2000/svg', 'text');

			textElement.setAttribute('stroke', fColor);
			textElement.setAttribute('text-anchor', 'middle');
			textElement.setAttribute('dy', '.3em');
			textElement.setAttribute('stroke-width', '2px');
			textElement.setAttribute('font-size', '25px');
			textElement.textContent = id;
			circleElement.setAttribute('r', 25); // Radius, adjust as needed
			circleElement.setAttribute('fill', bColor); // Color, adjust as needed
			circleElement.setAttribute('stroke', '#111111'); // Color, adjust as needed
			carElement.appendChild(circleElement);
			carElement.appendChild(textElement);

			svgContainer.appendChild(carElement);

			return {circle: circleElement, text: textElement, car: carElement};
		}

    // Function to draw a circle on the SVG
    function drawCar(carElements, x, y) {
			const circleElement = carElements.circle;
			const textElement = carElements.text;
			const carElement = carElements.car;

			textElement.setAttribute('x', x);
			textElement.setAttribute('y', y);
			circleElement.setAttribute('cx', x);
      circleElement.setAttribute('cy', y);
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

// var homeTemplate = template.Must(template.New("").Parse(`
// <!DOCTYPE html>
// <html>
// <head>
// <meta charset="utf-8">
// <script>
// window.addEventListener("load", function(evt) {
//     var output = document.getElementById("output");
//     var input = document.getElementById("input");
//     var ws;

//     var drawPoint = function(message) {
// 				console.log(message)
// 				var jsonObject = JSON.parse(message);
// 				var x = jsonObject.x;
// 				var z = jsonObject.z + 100;
// 				const circleElement = document.createElementNS('http://www.w3.org/2000/svg', 'circle');
// 				circleElement.setAttribute('cx', x); // x-coordinate of the center
// 				circleElement.setAttribute('cy', z); // y-coordinate of the center
// 				circleElement.setAttribute('r', 10);   // radius
// 				circleElement.setAttribute('fill', 'red'); // fill color
// 				const svgContainer = document.getElementById('mySvg');
// 				while (svgContainer.firstChild) {
// 					svgContainer.removeChild(svgContainer.firstChild);
// 			  }
// 				svgContainer.appendChild(circleElement);
//     };

// 		var print = function(message) {
// 			var d = document.createElement("div");
// 			d.textContent = message;
// 			output.appendChild(d);
// 			output.scroll(0, output.scrollHeight);
// 	};

//     document.getElementById("open").onclick = function(evt) {
//         if (ws) {
//             return false;
//         }
//         ws = new WebSocket("{{.}}");
//         ws.onopen = function(evt) {
// 				  console.log("OPEN");
//           print(evt);
//         }
//         ws.onclose = function(evt) {
// 					console.log("CLOSE");
//             print("CLOSE");
//             ws = null;
//         }
//         ws.onmessage = function(evt) {
//             console.log(evt.data)
//             drawPoint(evt.data);
//         }
//         ws.onerror = function(evt) {
// 						console.log("ERROR");
//             print("ERROR: " + evt.data);
//         }
//         return false;
//     };

//     document.getElementById("send").onclick = function(evt) {
//         if (!ws) {
//             return false;
//         }
//         print("SEND: " + input.value);
//         ws.send(input.value);
//         return false;
//     };

//     document.getElementById("close").onclick = function(evt) {
//         if (!ws) {
//             return false;
//         }
//         ws.close();
//         return false;
//     };

// });
// </script>
// </head>
// <body>
// <table>
// <tr><td valign="top" width="50%">
// <p>Click "Open" to create a connection to the server,
// "Send" to send a message to the server and "Close" to close the connection.
// You can change the message and send multiple times.
// <p>
// <form>
// <button id="open">Open</button>
// <button id="close">Close</button>
// <p><input id="input" type="text" value="Hello world!">
// <button id="send">Send</button>
// </form>
// <div width="500" height="500" id="svgContainer"></div>
// </td><td valign="top" width="50%">
// <div id="output" style="max-height: 70vh;overflow-y: scroll;"></div>
// <svg id="mySvg" width="1000" height="1000" xmlns="http://www.w3.org/2000/svg"></svg>
// <object type="image/svg+xml" data="http://localhost:8080/track/imola.svg" width="100%" height="100%"></object>
// </td></tr></table>
// </body>
// </html>
// `))

package main
import (
	//"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"os"
	"os/signal"
	"context"
	//"strconv"
	"syscall"

	"golang.org/x/net/websocket"
)

type File struct {name string; echo []byte; mode os.FileMode;}
type FileInterface interface { write() }
func (f File) write() { os.WriteFile(f.name, f.echo, f.mode) }

func dataSocket(stringchan chan string, file File) {

	var err error
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("[HTTP Server][Data Socket][Listen()]", err)
	}
	addr := listener.Addr().String()
	file.echo = []byte(addr[strings.LastIndex(addr, ":"):])
	file.write()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("[HTTP Server][Data Socket][Accept()]", err)
		}
		//defer listener.Close() /* doing nothing */
		const readBytes = 0x1000
		recieved := ""
		for func (recieved *string) error {
			var read [readBytes]byte
			readed, err := conn.Read(read[:])
			if err != nil {
				//log.Fatal("[Server Data Read]", err)
			}
			*recieved += string(read[:readed])
			return err
		}(&recieved) == nil {}
		stringchan <- recieved
	}

}

func getVimInfo() (vimPreviewRoot string, pid string, bufnr string) {
	b := make([]byte, 1)
	for func () int {
		readed, err := os.Stdin.Read(b)
		if b[0] == 0xff || err != nil {return -1}
		return readed
	}() > 0 {
		vimPreviewRoot += string(b)
	}
	for func () int {
		readed, err := os.Stdin.Read(b)
		if b[0] == 0xff || err != nil {return -1}
		return readed
	}() > 0 {
		pid += string(b)
	}
	for func () int {
		readed, err := os.Stdin.Read(b)
		if b[0] == 0xff || err != nil {return -1}
		return readed
	}() > 0 {
		bufnr += string(b)
	}
	return
}

func main() {

	vimPreviewRoot, vimPid, vimBufnr := getVimInfo()
	dataPort := File{
		name : vimPreviewRoot + "/" + vimPid + "." + vimBufnr + "." + "data.port",
		echo : []byte("0"),
		mode : 0644,
	}
	stringchan := make(chan string)
	go dataSocket(stringchan, dataPort)
	defer func () {
		err := os.Remove(dataPort.name)
		if err != nil {
			log.Printf("[HTTP Server][defer][remove file]: %s", dataPort.name)
		}
	}()

	http.HandleFunc("/", func (w http.ResponseWriter, req *http.Request) {
		wsServer := websocket.Server{Handler: websocket.Handler(
			func (wsConn *websocket.Conn) {
				beforeunload := make(chan byte)
				go func () {
					for {
						var received string
						if
						err := websocket.Message.Receive(wsConn, &received);
						err != nil {break}
						log.Printf("%v\n", received)
					}
					beforeunload <- 0
				}()
				go func () {
					for {
						fromDataSocket := <- stringchan
						if fromDataSocket == "" {continue}
						if
						err := websocket.Message.Send(wsConn, fromDataSocket)
						err != nil {
							stringchan <- fromDataSocket
							/* keep the failed send to after refresh browser */
							break
						}
					}
				}()
				<- beforeunload
			})}
		wsServer.ServeHTTP(w, req)
	})
	/* this will not allow any origin client
		http.Handle("/", websocket.Handler(func (wsConn_p *websocket.Conn) {
			var err error
			if func (err_p *error) error {
				*err_p = websocket.Message.Send(wsConn_p, "<h1>WEBSOCKET</h1>")
				return *err_p
			}(&err) != nil {
				log.Fatal("[Server]:[Websocket]:[Send]:", err)
			}
		}))
	*/

	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatal("[Server]:[TCP]:[Listrrn]:", err)
	}
	defer listener.Close() /* doing nothing */
	addr := listener.Addr().String()

	files := []File{
		{
			name : vimPreviewRoot + "/" + "vim" + "." + "pid",
			//echo : []byte(strconv.Itoa(os.Getpid())),
			echo : []byte(vimPid),
			mode : 0644,
		},
		{
			name : vimPreviewRoot + "/" + vimPid + "." + vimBufnr + "." + "websocket.port",
			echo : []byte(addr[strings.LastIndex(addr, ":"):]),
			mode : 0644,
		},
		{
			name : vimPreviewRoot + "/" + vimPid + "." + vimBufnr + "." + "websocket.html",
			echo : []byte(`
				<html><head>
				<script type="text/javascript">
					var websocket = null;
					var wsuri = "ws://wsl.localhost` + addr[strings.LastIndex(addr, ":"):] + `";
					window.onload = function () {
						console.log("onload");
						websocket = new WebSocket(wsuri);
						websocket.onopen = function () {};
						websocket.onmessage = function (e) {
							//location.assign(location.href.substr(0,location.href.indexOf('?')) + "?randomForRefresh=" + Date.now())
							const body = document.getElementsByTagName('body');
							body[0].innerHTML = e.data;
							//var imgs = document.getElementsByTagName('img');
							//imgs = Array.prototype.slice.call(imgs); /* DOM list -> DOM node array */
							//imgs.forEach((val) => refresh(val));

							//var svgs = document.getElementsByTagName('svg');
							//svgs = Array.prototype.slice.call(svgs);
							//svgs.forEach((val) => val.removeAttribute("viewbox"));
							//if(window.innerHeight < window.innerWidth) {
							//	svgs.forEach((val) => val.setAttribute("height", "100%"));
							//	svgs.forEach((val) => val.removeAttribute("width"));
							//} else {
							//	svgs.forEach((val) => val.setAttribute("width", "100%"));
							//	svgs.forEach((val) => val.removeAttribute("height"));
							//}
						};
					};
					window.onbeforeunload = function () {
						websocket.onclose = function () {};
						websocket.close();
					};
				</script></head>
				<body/>
				</html>
			`),
			mode : 0644,
		},
	}
	for _, f := range files {
		f.write()
	}
	defer func () {
		for _, f := range files {
			err := os.Remove(f.name)
			if err != nil {
				log.Printf("[HTTP Server][defer][remove file]: %s", f.name)
			}
		}
	}()

	var server http.Server
	idleConnsClosed := make(chan struct{})
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
		<-sigChan

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}
		close(idleConnsClosed)
	}()

	err = server.Serve(listener)
	if err != http.ErrServerClosed {
		log.Fatalf("[HTTP Server]:[Serve()]: %v", err)
	}

}

package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/pfedan/goPhoBo/src/randimg"

	"github.com/gorilla/mux"
	"github.com/looplab/fsm"
	"github.com/nfnt/resize"
	flag "github.com/spf13/pflag"
)

// PhoBoFlags defines the setings that can be set via command-line flags
type PhoBoFlags struct {
	nameOfParty string
	thumbWidth  uint
	port        string
	imgPath     string
	staticPath  string
}

// PhoBo structure, containing all needed stuff
type PhoBo struct {
	FSM           *fsm.FSM
	cntPhotos     uint64
	remoteCommand string
	lastImageName string
	f             *PhoBoFlags
}

// NewPhoBo is an initializer function for a PhoBo
func NewPhoBo(pFlags *PhoBoFlags) *PhoBo {
	d := &PhoBo{f: pFlags}

	d.FSM = fsm.NewFSM(
		"home",
		fsm.Events{
			{Name: "doPhoto", Src: []string{"home"}, Dst: "makingPhoto"},
			{Name: "beginDecide", Src: []string{"makingPhoto"}, Dst: "decide"},
			{Name: "deletePhoto", Src: []string{"decide"}, Dst: "home"},
			{Name: "acceptPhoto", Src: []string{"decide"}, Dst: "home"},
			{Name: "beginSmile", Src: []string{"home"}, Dst: "smile"},
			{Name: "endSmile", Src: []string{"smile"}, Dst: "home"},
		},
		fsm.Callbacks{
			"enter_state":  func(e *fsm.Event) { d.enterState(e) },
			"before_event": func(e *fsm.Event) { d.beforeEvent(e) },
			"doPhoto":      func(e *fsm.Event) { d.cbDoPhoto(e) },
			"deletePhoto":  func(e *fsm.Event) { d.cbDeletePhoto(e) },
			"acceptPhoto":  func(e *fsm.Event) { d.cbAcceptPhoto(e) },
			"beginSmile":   func(e *fsm.Event) { d.cbBeginSmile(e) },
			"endSmile":     func(e *fsm.Event) { d.cbEndSmile(e) },
		},
	)

	d.remoteCommand = "nothing"
	d.lastImageName = ""
	d.cntPhotos = uint64(len(getImageFileNames(d.f.imgPath)))

	return d
}

func (d *PhoBo) enterState(e *fsm.Event) {
	log.Printf("State changed from \"%s\" to \"%s\" \n", e.Src, e.Dst)
}

func (d *PhoBo) beforeEvent(e *fsm.Event) {
	log.Printf("New event: \"%s\"\n", e.Event)

}

func (d *PhoBo) cbDoPhoto(e *fsm.Event) {
	fname := time.Now().Format("2006-01-02T15-04-05.jpg")
	newpath := filepath.Clean(filepath.Join(d.f.imgPath, "small"))
	os.MkdirAll(newpath, os.ModePerm)
	o := jpeg.Options{Quality: 90}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("cmd", "/C", "echo I should have made a photo.").Output()
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Executed photo command. \n   -> Output: %s", out)

		m := randimg.GetImg(randimg.RandImgOptions{W: 800, H: 600})

		fa, erra := os.OpenFile(d.f.imgPath+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if erra != nil {
			fmt.Println(erra)
			return
		}
		defer fa.Close()
		jpeg.Encode(fa, m, &o)
		saveThumbnail(m, d, fname)
	} else {
		gphotoCmd := exec.Command("bash", "-c", "gphoto2 --auto-detect --capture-image-and-download --force-overwrite --filename "+d.f.imgPath+fname)
		out, err := gphotoCmd.CombinedOutput()
		if err != nil {
			log.Printf("Executed photo command. \n   -> FAILED: Output: \n%s", out)
			log.Fatal(err)
		}
		log.Printf("Executed photo command. \n   -> Output: \n%s", out)

		// open new photo
		fa, err := os.Open(d.f.imgPath + fname)
		if err != nil {
			log.Fatal(err)
		}
		defer fa.Close()

		m, _, err := image.Decode(fa)
		if err != nil {
			log.Fatal(err)
		}
		saveThumbnail(m, d, fname)
	}
	d.lastImageName = fname
}

func saveThumbnail(img image.Image, d *PhoBo, fname string) {
	if m, ok := img.(*image.RGBA); ok {
		o := jpeg.Options{Quality: 90}
		mThumbnail := resize.Resize(d.f.thumbWidth, 0, m, resize.Bicubic)
		fb, errb := os.OpenFile(d.f.imgPath+"small/"+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if errb != nil {
			fmt.Println(errb)
			return
		}
		defer fb.Close()
		jpeg.Encode(fb, mThumbnail, &o)
	}
}

func (d *PhoBo) cbDeletePhoto(e *fsm.Event) {
	err := os.Remove(d.f.imgPath + d.lastImageName)
	if err != nil {
		log.Println("Could not delete file:", d.f.imgPath+d.lastImageName)
	}
	err = os.Remove(d.f.imgPath + "small/" + d.lastImageName)
	if err != nil {
		log.Println("Could not delete file:", d.f.imgPath+"small/"+d.lastImageName)
	}

	d.lastImageName = ""
}

func (d *PhoBo) cbAcceptPhoto(e *fsm.Event) {
	d.cntPhotos++
}

func (d *PhoBo) cbBeginSmile(e *fsm.Event) {

}

func (d *PhoBo) cbEndSmile(e *fsm.Event) {

}

func (d *PhoBo) decideForMeAfter(t time.Duration) {
	time.Sleep(t)
	if rand.Float32() < 0.5 {
		d.FSM.Event("deletePhoto")
	} else {
		d.FSM.Event("acceptPhoto")
	}

}

func (d *PhoBo) emitEventAfter(e string, t time.Duration) {
	time.Sleep(t)
	d.FSM.Event(e)
}

type responseStatus struct {
	EventSuccess        bool     `json:"eventSuccess"`
	CntPhotos           uint64   `json:"cntPhotos"`
	CurrentState        string   `json:"currentState"`
	PossibleTransitions []string `json:"possibleTransitions"`
	RemoteCommand       string   `json:"remoteCommand"`
}

type eventRouteInfo struct {
	router    *mux.Router
	p         *PhoBo
	event     string
	route     string
	fPossible func(*PhoBo)
}

func registerEventRoute(o eventRouteInfo) {
	o.router.HandleFunc(o.route, func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, o)
	})
}

// handleEventRoute handles actions and response after a request
func handleEventRoute(w http.ResponseWriter, r *http.Request, o eventRouteInfo) {
	fsm := o.p.FSM
	possible := fsm.Can(o.event)

	if possible {
		o.fPossible(o.p)
	}

	res := responseStatus{
		EventSuccess:        possible,
		CntPhotos:           o.p.cntPhotos,
		CurrentState:        fsm.Current(),
		PossibleTransitions: fsm.AvailableTransitions(),
		RemoteCommand:       o.p.remoteCommand,
	}

	json.NewEncoder(w).Encode(res)
}

func getImageFileNames(path string) []string {
	var list []string
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return list
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}
		list = append(list, f.Name())
	}
	return list
}

func getStatus(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	if id := params["id"]; id != "" {
		if id == "remoteCommand" {
			mPhoBo.remoteCommand = params["val"]
		}
	}

	res := responseStatus{
		CntPhotos:           mPhoBo.cntPhotos,
		CurrentState:        mPhoBo.FSM.Current(),
		PossibleTransitions: mPhoBo.FSM.AvailableTransitions(),
		RemoteCommand:       mPhoBo.remoteCommand,
	}

	json.NewEncoder(w).Encode(res)
}

var flagPhoBo PhoBoFlags
var mPhoBo *PhoBo

func init() {
	const (
		portDefault   = "8080"
		portUsage     = "port for web-frontend"
		nameDefault   = "PhoBo-Party"
		nameUsage     = "Name of the Event"
		imgDefault    = "img/"
		imgUsage      = "Path to image files"
		staticDefault = "static/"
		staticUsage   = "Path to static webserver files"
		thumbDefault  = 240
		thumbUsage    = "Width of image thumbnails in px"
	)
	flag.StringVarP(&(flagPhoBo.port), "port", "p", portDefault, portUsage)
	flag.StringVarP(&(flagPhoBo.nameOfParty), "name", "n", nameDefault, nameUsage)
	flag.StringVarP(&(flagPhoBo.imgPath), "imgpath", "i", imgDefault, imgUsage)
	flag.StringVarP(&(flagPhoBo.staticPath), "staticpath", "s", staticDefault, staticUsage)
	flag.UintVarP(&(flagPhoBo.thumbWidth), "thumbnail-width", "t", thumbDefault, thumbUsage)
}

func initEventRoutes(p *PhoBo, r *mux.Router) {
	registerEventRoute(eventRouteInfo{router: r, route: "/doPhoto", event: "doPhoto", p: p, fPossible: func(p *PhoBo) {
		p.FSM.Event("doPhoto")
		p.FSM.Event("beginDecide")
		//go p.decideForMeAfter(1 * time.Second)
	}})

	registerEventRoute(eventRouteInfo{router: r, route: "/deletePhoto", event: "deletePhoto", p: p, fPossible: func(p *PhoBo) {
		p.FSM.Event("deletePhoto")
	}})

	registerEventRoute(eventRouteInfo{router: r, route: "/acceptPhoto", event: "acceptPhoto", p: p, fPossible: func(p *PhoBo) {
		p.FSM.Event("acceptPhoto")
	}})

	registerEventRoute(eventRouteInfo{router: r, route: "/beginSmile", event: "beginSmile", p: p, fPossible: func(p *PhoBo) {
		p.FSM.Event("beginSmile")
		go p.emitEventAfter("endSmile", 3*time.Second)
	}})

	registerEventRoute(eventRouteInfo{router: r, route: "/endSmile", event: "endSmile", p: p, fPossible: func(p *PhoBo) {
		p.FSM.Event("endSmile")
	}})
}

func main() {
	flag.Parse()
	fmt.Printf("%+v\n", flagPhoBo)

	mPhoBo = NewPhoBo(&flagPhoBo)
	router := mux.NewRouter()

	initEventRoutes(mPhoBo, router)

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(mPhoBo.f.staticPath))))
	router.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir(mPhoBo.f.imgPath))))

	router.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string][]string{
			"imageFiles": getImageFileNames(mPhoBo.f.imgPath),
		})
	})

	router.HandleFunc("/status", getStatus)
	router.HandleFunc("/status/{id}/{val}", getStatus).Methods("GET")

	router.HandleFunc("/deleteAll", func(w http.ResponseWriter, r *http.Request) {
		n := len(getImageFileNames(mPhoBo.f.imgPath))
		os.RemoveAll(mPhoBo.f.imgPath)
		mPhoBo.cntPhotos = 0
		log.Printf("All %v images in %s have been deleted.\n", n, mPhoBo.f.imgPath)
	})

	router.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"commandSuccess": true})
		os.Exit(0)
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + mPhoBo.f.port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

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
	"github.com/k0kubun/go-ansi"
	"github.com/looplab/fsm"
	"github.com/nfnt/resize"
	flag "github.com/spf13/pflag"
)

// PhoBo structure, containing all needed stuff
type PhoBo struct {
	FSM        *fsm.FSM
	cntPhotos  uint64
	smallWidth uint
}

type phoBoFlags struct {
	nameOfParty string
	smallWidth  uint
	port        string
	imgPath     string
	staticPath  string
}

// NewPhoBo is an initializer function for a PhoBo
func NewPhoBo() *PhoBo {
	d := &PhoBo{}

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

	d.cntPhotos = uint64(len(getImageFileNames(flagPhoBo.imgPath)))
	d.smallWidth = 240

	return d
}

func (d *PhoBo) enterState(e *fsm.Event) {
	fmt.Printf("State changed from \"%s\" to \"%s\" \n", e.Src, e.Dst)
}

func (d *PhoBo) beforeEvent(e *fsm.Event) {
	fmt.Printf("New event: \"%s\"\n", e.Event)

}

func (d *PhoBo) cbDoPhoto(e *fsm.Event) {
	for i := 3.0; i > 0; i -= 1 {
		ansi.CursorHorizontalAbsolute(1)
		ansi.EraseInLine(0)
		fmt.Printf("Countdown: %.1f\n", i)
		time.Sleep(1000 * time.Millisecond)
	}

	fname := time.Now().Format("2006-01-02T15-04-05.jpg")
	newpath := filepath.Clean(filepath.Join(flagPhoBo.imgPath, "small"))
	os.MkdirAll(newpath, os.ModePerm)

	o := jpeg.Options{Quality: 90}
	if runtime.GOOS == "windows" {
		out, err := exec.Command("cmd", "/C", "echo I should have made a photo.").Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   -> executed photo command. \n   -> Output: %s", out)

		m := randimg.GetImg(randimg.RandImgOptions{W: 800, H: 600})

		fa, erra := os.OpenFile(flagPhoBo.imgPath+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if erra != nil {
			fmt.Println(erra)
			return
		}
		defer fa.Close()
		jpeg.Encode(fa, m, &o)

		// also save a thumbnail
		mThumbnail := resize.Resize(d.smallWidth, 0, m, resize.Bicubic)
		fb, errb := os.OpenFile(flagPhoBo.imgPath+"small/"+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if errb != nil {
			fmt.Println(errb)
			return
		}
		defer fb.Close()
		jpeg.Encode(fb, mThumbnail, &o)
	} else {
		gphotoCmd := exec.Command("bash", "-c", "gphoto2 --auto-detect --capture-image-and-download --force-overwrite --filename "+flagPhoBo.imgPath+fname)
		fmt.Println(gphotoCmd)
		out, err := gphotoCmd.Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   -> executed photo command. \n   -> Output: \n%s", out)

		// open new photo
		fa, err := os.Open(flagPhoBo.imgPath + fname)
		if err != nil {
			log.Fatal(err)
		}
		defer fa.Close()

		m, _, err := image.Decode(fa)
		if err != nil {
			log.Fatal(err)
		}

		mThumbnail := resize.Resize(d.smallWidth, 0, m, resize.Bicubic)
		fb, errb := os.OpenFile(flagPhoBo.imgPath+"small/"+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if errb != nil {
			fmt.Println(errb)
			return
		}
		defer fb.Close()
		jpeg.Encode(fb, mThumbnail, &o)
	}

	d.cntPhotos++
}

func (d *PhoBo) cbDeletePhoto(e *fsm.Event) {

}

func (d *PhoBo) cbAcceptPhoto(e *fsm.Event) {

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

func (d *PhoBo) listEventLinks() string {
	var ret string
	for _, v := range d.FSM.AvailableTransitions() {
		ret += "<a href=\"" + v + "\">" + v + "</a> "
	}
	return ret
}

type responseStatus struct {
	EventSuccess        bool     `json:"eventSuccess"`
	CntPhotos           uint64   `json:"cntPhotos"`
	CurrentState        string   `json:"currentState"`
	PossibleTransitions []string `json:"possibleTransitions"`
}

// handleEventRoute handles actions and response after a request
func handleEventRoute(w http.ResponseWriter, r *http.Request, p *PhoBo, e string, fPossible func(*PhoBo)) {
	possible := p.FSM.Can(e)

	if possible {
		fPossible(p)
	}

	res := responseStatus{
		EventSuccess:        possible,
		CntPhotos:           p.cntPhotos,
		CurrentState:        p.FSM.Current(),
		PossibleTransitions: p.FSM.AvailableTransitions(),
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

var flagPhoBo phoBoFlags

func init() {
	const (
		portDefault   = "http"
		portUsage     = "port for web-frontend"
		nameDefault   = "PhoBo-Party"
		nameUsage     = "Name of the Event"
		imgDefault    = "img/"
		imgUsage      = "Path to image files"
		staticDefault = "static/"
		staticUsage   = "Path to static webserver files"
	)
	flag.StringVarP(&(flagPhoBo.port), "port", "p", portDefault, portUsage)
	flag.StringVarP(&(flagPhoBo.nameOfParty), "name", "n", nameDefault, nameUsage)
	flag.StringVarP(&(flagPhoBo.imgPath), "imgpath", "i", imgDefault, imgUsage)
	flag.StringVarP(&(flagPhoBo.staticPath), "staticpath", "s", staticDefault, staticUsage)
}

func main() {
	flag.Parse()
	fmt.Printf("%+v\n", flagPhoBo)
	mPhoBo := NewPhoBo()
	router := mux.NewRouter()

	router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(flagPhoBo.staticPath))))
	router.PathPrefix("/img/").Handler(http.StripPrefix("/img/", http.FileServer(http.Dir(flagPhoBo.imgPath))))

	router.HandleFunc("/doPhoto", func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, mPhoBo, "doPhoto",
			func(p *PhoBo) {
				p.FSM.Event("doPhoto")
				p.FSM.Event("beginDecide")
				go p.decideForMeAfter(1 * time.Second)
			})
	})

	router.HandleFunc("/deletePhoto", func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, mPhoBo, "deletePhoto",
			func(p *PhoBo) {
				p.FSM.Event("deletePhoto")
			})
	})

	router.HandleFunc("/acceptPhoto", func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, mPhoBo, "acceptPhoto",
			func(p *PhoBo) {
				p.FSM.Event("acceptPhoto")
			})
	})

	router.HandleFunc("/beginSmile", func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, mPhoBo, "beginSmile",
			func(p *PhoBo) {
				p.FSM.Event("beginSmile")
				go p.emitEventAfter("endSmile", 3*time.Second)
			})
	})

	router.HandleFunc("/endSmile", func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, mPhoBo, "endSmile",
			func(p *PhoBo) {
				p.FSM.Event("endSmile")
			})
	})

	router.HandleFunc("/images", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string][]string{
			"imageFiles": getImageFileNames(flagPhoBo.imgPath),
		})
	})

	router.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		res := responseStatus{
			CntPhotos:           mPhoBo.cntPhotos,
			CurrentState:        mPhoBo.FSM.Current(),
			PossibleTransitions: mPhoBo.FSM.AvailableTransitions(),
		}

		json.NewEncoder(w).Encode(res)
	})

	router.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"commandSuccess": true})
		os.Exit(0)
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         ":" + flagPhoBo.port,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

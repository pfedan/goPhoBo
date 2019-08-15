package main

import (
	"encoding/json"
	"fmt"
	"goPhoBo/src/randimg"
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

	"github.com/gorilla/mux"
	"github.com/k0kubun/go-ansi"
	"github.com/looplab/fsm"
	"github.com/nfnt/resize"
)

// PhoBo structure, containing all needed stuff
type PhoBo struct {
	nameOfParty string
	FSM         *fsm.FSM
	cntPhotos   uint64
	smallWidth  uint
}

// NewPhoBo is an initializer function for a PhoBo
func NewPhoBo(newNameOfParty string) *PhoBo {
	d := &PhoBo{
		nameOfParty: newNameOfParty,
	}

	d.FSM = fsm.NewFSM(
		"home",
		fsm.Events{
			{Name: "doPhoto", Src: []string{"home"}, Dst: "decide"},
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

	d.cntPhotos = 0
	d.smallWidth = 120

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

	d.cntPhotos++

	if runtime.GOOS == "windows" {
		out, err := exec.Command("cmd", "/C", "echo I should have made a photo.").Output()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("   -> executed photo command. \n   -> Output: %s", out)

		m := randimg.GetImg(800, 600)

		newpath := filepath.Join(".", "img", "small")
		os.MkdirAll(newpath, os.ModePerm)

		fname := time.Now().Format("2006-01-02T15-04-05.jpg")
		fa, erra := os.OpenFile("img/"+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if erra != nil {
			fmt.Println(err)
			return
		}
		defer fa.Close()
		o := jpeg.Options{Quality: 90}
		jpeg.Encode(fa, m, &o)

		// also save a thumbnail
		mThumbnail := resize.Resize(d.smallWidth, 0, m, resize.Bicubic)
		fb, errb := os.OpenFile("img/small/"+fname, os.O_WRONLY|os.O_CREATE, 0600)
		if errb != nil {
			fmt.Println(errb)
			return
		}
		defer fb.Close()
		jpeg.Encode(fb, mThumbnail, &o)

	}

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

// handleEventRoute handles actions and response after a request
func handleEventRoute(w http.ResponseWriter, r *http.Request, p *PhoBo, e string, fPossible func(*PhoBo)) {
	possible := p.FSM.Can(e)

	if possible {
		fPossible(p)
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"eventSuccess":        possible,
		"cntPhotos":           p.cntPhotos,
		"currentState":        p.FSM.Current(),
		"possibleTransitions": p.FSM.AvailableTransitions(),
	})
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

func main() {
	mPhoBo := NewPhoBo("Party")

	router := mux.NewRouter()

	router.HandleFunc("/doPhoto", func(w http.ResponseWriter, r *http.Request) {
		handleEventRoute(w, r, mPhoBo, "doPhoto",
			func(p *PhoBo) {
				p.FSM.Event("doPhoto")
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
		json.NewEncoder(w).Encode(map[string]interface{}{
			"imageFiles": getImageFileNames("img/"),
		})
	})

	router.HandleFunc("/quit", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"commandSuccess": true})
		os.Exit(0)
	})

	srv := &http.Server{
		Handler:      router,
		Addr:         "127.0.0.1:8000",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}

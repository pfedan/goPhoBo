package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/looplab/fsm"
)

// PhoBo structure, containing all needed stuff
type PhoBo struct {
	nameOfParty string
	FSM         *fsm.FSM
	cntPhotos   uint64
}

var pPhoBo *PhoBo

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

	return d
}

func (d *PhoBo) enterState(e *fsm.Event) {
	fmt.Printf("State changed from \"%s\" to \"%s\" \n", e.Src, e.Dst)
}

func (d *PhoBo) beforeEvent(e *fsm.Event) {
	fmt.Printf("New event: \"%s\"\n", e.Event)

}

func (d *PhoBo) cbDoPhoto(e *fsm.Event) {
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

func (d *PhoBo) endSmileAfter(t time.Duration) {
	time.Sleep(t)
	d.FSM.Event("endSmile")
}

func (d *PhoBo) decideForMeAfter(t time.Duration) {
	time.Sleep(t)
	if rand.Float32() < 0.5 {
		d.FSM.Event("deletePhoto")
	} else {
		d.FSM.Event("acceptPhoto")
	}

}

func main() {
	mPhoBo := NewPhoBo("Party")
	pPhoBo = mPhoBo

	router := mux.NewRouter()

	router.HandleFunc("/doPhoto", func(w http.ResponseWriter, r *http.Request) {
		pPhoBo.FSM.Event("doPhoto")
		go pPhoBo.decideForMeAfter(1 * time.Second)
		json.NewEncoder(w).Encode(map[string]bool{"eventSuccess": true})
		json.NewEncoder(w).Encode(map[string]uint64{"cntPhotos": pPhoBo.cntPhotos})
	})

	router.HandleFunc("/deletePhoto", func(w http.ResponseWriter, r *http.Request) {
		pPhoBo.FSM.Event("deletePhoto")
		json.NewEncoder(w).Encode(map[string]bool{"eventSuccess": true})
	})

	router.HandleFunc("/acceptPhoto", func(w http.ResponseWriter, r *http.Request) {
		pPhoBo.FSM.Event("acceptPhoto")
		json.NewEncoder(w).Encode(map[string]bool{"eventSuccess": true})
	})

	router.HandleFunc("/beginSmile", func(w http.ResponseWriter, r *http.Request) {
		pPhoBo.FSM.Event("beginSmile")
		go pPhoBo.endSmileAfter(3 * time.Second)
		json.NewEncoder(w).Encode(map[string]bool{"eventSuccess": true})
	})

	router.HandleFunc("/endSmile", func(w http.ResponseWriter, r *http.Request) {
		pPhoBo.FSM.Event("endSmile")
		json.NewEncoder(w).Encode(map[string]bool{"eventSuccess": true})
	})

	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())

	// mPhoBo.FSM.Event("doPhoto")
	// mPhoBo.FSM.Event("acceptPhoto")
	// mPhoBo.FSM.Event("doPhoto")
	// mPhoBo.FSM.Event("deletePhoto")
	// mPhoBo.FSM.Event("beginSmile")
	// mPhoBo.FSM.Event("endSmile")

	// time.Sleep(3 * time.Second)
}

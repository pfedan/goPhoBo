package main

import (
	"net/http"
	"testing"
	"time"

	"github.com/gorilla/mux"
	flag "github.com/spf13/pflag"
)

func Test_initEventRoutes(t *testing.T) {
	flag.Parse()
	mPhoBo := NewPhoBo(&flagPhoBo)
	router := mux.NewRouter()

	type args struct {
		p *PhoBo
		r *mux.Router
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "default",
			args: args{
				p: mPhoBo,
				r: router,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initEventRoutes(tt.args.p, tt.args.r)
		})
	}
}

func Test_fsm(t *testing.T) {
	flag.Parse()
	p := NewPhoBo(&flagPhoBo)

	t.Run("RunAllStates", func(t *testing.T) {
		p.FSM.Event("doPhoto")
		p.FSM.Event("beginDecide")
		p.FSM.Event("deletePhoto")
		p.FSM.Event("doPhoto")
		p.FSM.Event("beginDecide")
		p.FSM.Event("acceptPhoto")
		p.FSM.Event("beginSmile")
		p.FSM.Event("endSmile")
	})
}

func Test_routes(t *testing.T) {
	go main()

	time.Sleep(1 * time.Second)

	tests := []struct {
		name string
	}{
		{name: "beginSmile"},
		{name: "endSmile"},
		{name: "doPhoto"},
		{name: "acceptPhoto"},
		{name: "doPhoto"},
		{name: "deletePhoto"},
		{name: "status"},
		{name: "deleteAll"},
		{name: "images"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := http.Get("http://localhost:8080/" + tt.name)
			time.Sleep(350 * time.Millisecond)
			if err != nil || res.Status != "200 OK" {
				t.Errorf("EventRoute \"%s\" did not respond.", tt.name)
			}
		})
	}
}

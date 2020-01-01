package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"zenhack.net/go/sandstorm/exp/websession"
	"zenhack.net/go/sandstorm/grain"

	graincp "zenhack.net/go/sandstorm/capnp/grain"

	"zombiezen.com/go/capnproto2"
)

func chkfatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

type mainView struct {
	*websession.HandlerUiView
}

var _ graincp.MainView_Server = mainView{}

func (mainView) GetViewInfo(ctx context.Context, p graincp.UiView_getViewInfo) error {
	return nil
}

func (mainView) Restore(ctx context.Context, p graincp.MainView_restore) error {
	log.Println("Called restore")

	objectIdRaw, err := p.Args().ObjectId()
	chkfatal(err)
	objectId := AppObjectId{Struct: objectIdRaw.Struct()}
	name, err := objectId.CallbackName()
	chkfatal(err)

	ret := someCallback{name: name}

	res, err := p.AllocResults()
	chkfatal(err)
	cb := graincp.ScheduledJob_Callback_ServerToClient(ret, nil)
	capId := res.Struct.Segment().Message().AddCap(cb.Client)
	iface := capnp.NewInterface(res.Struct.Segment(), capId)
	res.SetCap(iface.ToPtr())

	return nil
}

func (mainView) Drop(ctx context.Context, p graincp.MainView_drop) error {
	return nil
}

type someCallback struct {
	name string
}

func (s someCallback) Run(ctx context.Context, req graincp.ScheduledJob_Callback_run) error {
	log.Println("Ran callback:", s.name)
	return nil
}

func (s someCallback) Save(ctx context.Context, req graincp.AppPersistent_save) error {
	log.Println("Saving:", s)
	res, err := req.AllocResults()
	chkfatal(err)

	label, err := res.NewLabel()
	chkfatal(err)
	label.SetDefaultText("Scheduled Job: " + s.name)

	objectId, err := NewAppObjectId(res.Segment())
	chkfatal(err)

	objectId.SetCallbackName(s.name)
	res.SetObjectId(objectId.ToPtr())

	return nil
}

func scheduleOnGet(name string, setSchedule func(p graincp.ScheduledJob)) func(context.Context) {
	return func(ctx context.Context) {
		api, err := grain.GetAPI()
		chkfatal(err)
		api.Schedule(ctx, func(p graincp.ScheduledJob) error {
			log.Printf("Scheduling job (%s)", name)
			nameL10n, err := p.NewName()
			chkfatal(err)
			nameL10n.SetDefaultText(name)

			cb := graincp.ScheduledJob_Callback{
				Client: AppPersistentCallback_ServerToClient(
					someCallback{name: name},
					nil,
				).Client,
			}
			p.SetCallback(cb)
			setSchedule(p)
			return nil
		})
	}
}

func noOpOnGet(ctx context.Context) {
	log.Println("Continuing; not scheduling a new job.")
}

func main() {
	var onGet func(context.Context)
	switch os.Args[1] {
	case "hourly":
		onGet = scheduleOnGet("hourly", func(p graincp.ScheduledJob) {
			p.Schedule().SetPeriodic(graincp.SchedulingPeriod_hourly)
		})
	case "oneShot":
		onGet = scheduleOnGet("oneShot", func(p graincp.ScheduledJob) {
			sched := p.Schedule()
			sched.SetOneShot()
			o := sched.OneShot()
			o.SetWhen(time.Now().Add(time.Second * 5).UnixNano())
		})
	case "continue":
		onGet = noOpOnGet
	}
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		onGet(req.Context())
		w.Write([]byte("Check the debug log"))
	})

	h := mainView{&websession.HandlerUiView{http.DefaultServeMux}}
	panic(websession.ListenAndServe(nil, h, nil))
}

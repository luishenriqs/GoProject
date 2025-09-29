package main

import "net/http"

func main() {
    mux := http.NewServeMux()
	mux2 := http.NewServeMux()

    mux.HandleFunc("/", SchedulerHandler)
    mux2.HandleFunc("/", AppointmentHandler)


	go func() {
		http.ListenAndServe(":8080", mux)
	}()

	http.ListenAndServe(":8081", mux2)

}

func SchedulerHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Scheduler Handler"))
}

func AppointmentHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Appointment Handler"))
}



package portapp_test

import (
	"testing"
	"time"

	"arcticdispatch/internal/domain"
	"arcticdispatch/internal/portapp"
	"arcticdispatch/internal/store"
)

// TestRescheduleKeepsPickupWindowConsistent re-appoints a truck long after its
// original pickup window has passed. The new window must still be a usable
// window: it starts no earlier than the moment of re-appointment, ends after it
// starts, and keeps its original length.
func TestRescheduleKeepsPickupWindowConsistent(t *testing.T) {
	st, clk, svc := newPortappEnv(t)
	st.Update(func(d *store.Data) error {
		d.Bookings["BK-1"] = &domain.Booking{
			ID: "BK-1", VoyageID: "V1", CargoID: "C1",
			Status: domain.BookingConfirmed, Priority: 100, SubmitTime: clk.Now(),
		}
		return nil
	})

	ws := clk.Now().Add(time.Hour)
	we := ws.Add(time.Hour)
	apt, err := svc.CreateAppointment(portapp.CreateAppointmentReq{
		BookingID: "BK-1", TruckID: "T1", WindowStart: ws, WindowEnd: we,
	})
	if err != nil {
		t.Fatal(err)
	}

	// The ship is delayed; re-appointment happens 5h after the original start.
	clk.Advance(6 * time.Hour)
	ids := svc.RescheduleForVoyage("V1")
	if len(ids) != 1 || ids[0] != apt.ID {
		t.Fatalf("want [%s] rescheduled, got %v", apt.ID, ids)
	}

	got, ok := svc.Get(apt.ID)
	if !ok {
		t.Fatalf("appointment %s missing after reschedule", apt.ID)
	}
	if got.Status != domain.AptRescheduled {
		t.Fatalf("want rescheduled, got %s", got.Status)
	}
	if !got.WindowEnd.After(got.WindowStart) {
		t.Fatalf("rescheduled window must end after it starts, got start=%s end=%s",
			got.WindowStart.Format(time.RFC3339), got.WindowEnd.Format(time.RFC3339))
	}
	if d := got.WindowEnd.Sub(got.WindowStart); d != time.Hour {
		t.Fatalf("rescheduled window length = %v, want the original 1h", d)
	}
	if got.WindowStart.Before(clk.Now()) {
		t.Fatalf("rescheduled window must not start in the past: start=%s now=%s",
			got.WindowStart.Format(time.RFC3339), clk.Now().Format(time.RFC3339))
	}
}

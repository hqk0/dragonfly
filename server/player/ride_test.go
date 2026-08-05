package player

import (
	"testing"

	"github.com/df-mc/dragonfly/server/entity"
	"github.com/df-mc/dragonfly/server/world"
	"github.com/go-gl/mathgl/mgl64"
)

func TestMountAndDismount(t *testing.T) {
	w := (world.Config{Synchronous: true}).New()
	defer w.Close()

	seatOffset := mgl64.Vec3{0, 0.9, 2}
	w.Do(func(tx *world.Tx) {
		playerHandle := (world.EntitySpawnOpts{
			Position: mgl64.Vec3{0, 1, 0},
		}).New(Type, Config{Name: "rider"})
		vehicleHandle := entity.NewText("vehicle", mgl64.Vec3{0, 1, 0})
		rider := tx.AddEntity(playerHandle).(*Player)
		vehicle := tx.AddEntity(vehicleHandle)

		if rider.Silent() {
			t.Fatal("player was silent before mounting")
		}
		if !rider.Mount(vehicle, seatOffset) {
			t.Fatal("player could not mount vehicle")
		}
		if !rider.Silent() {
			t.Fatal("mounted player was not silent")
		}
		got, riding := rider.RiderSeatOffset()
		if !riding || got != seatOffset {
			t.Fatalf("seat offset = %v, riding=%v", got, riding)
		}

		rider.Dismount()
		if _, riding := rider.RiderSeatOffset(); riding {
			t.Fatal("player remained mounted after dismount")
		}
		if rider.Silent() {
			t.Fatal("player remained silent after dismount")
		}
	})
}

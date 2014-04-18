package game

import (
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

type EntityManager struct {
	engine *ecs.Engine
}

func NewEntityManager(e *ecs.Engine) *EntityManager {
	return &EntityManager{
		engine: e,
	}
}

func (m *EntityManager) Initalize() {
	s := ecs.NewEntity(
		"game",
		&GameStateComponent{"init", time.Now()},
	)

	if err := m.engine.AddEntity(s); err != nil {
		log.Fatal(err)
	}
}

func (m *EntityManager) CreateSplashScreen() {}

func (m *EntityManager) CreateMainMenu() {}

var (
	deg2rad = math.Pi / 180.0

	MaxShipSpeed      = 100.0 // pixels per second
	MaxAccelerate     = MaxShipSpeed
	ShipRotationSpeed = 180.0 // deg per second

	TimeBetweenBullets = 250 * time.Millisecond
	BulletSpeed        = 2 * MaxShipSpeed
	BulletLifetime     = 5 * time.Second

	MaxAsteroidRotation = 2 * ShipRotationSpeed
	MaxAsteroidSpeed    = MaxShipSpeed
)

func (em *EntityManager) CreateAsteroid(x, y float64, size int) {
	rot := rand.Float64() * 360
	rad := (rot + 90) * deg2rad
	speed := rand.Float64() * MaxAsteroidSpeed

	a := ecs.NewEntity(
		"asteroid",

		&PositionComponent{Point{x, y}, rot},
		&VelocityComponent{
			Point{
				speed * math.Cos(rad),
				speed * math.Sin(rad),
			}, MaxAsteroidRotation * rand.Float64(),
		},

		&ColorComponent{1, 1, 0},
	)

	mc := &MeshComponent{
		Points: make([]Point, 7),
		Max:    0,
	}

	step := (2.0 * math.Pi) / float64(len(mc.Points))
	max := float64(size * 10)
	min := max / 2

	for i := range mc.Points {
		length := (rand.Float64() * (max - min)) + min
		angle := float64(i) * step

		mc.Points[i].X = length * math.Cos(angle)
		mc.Points[i].Y = length * math.Sin(angle)

		mc.Max = math.Max(mc.Max, length)
	}

	a.Add(mc)

	if err := em.engine.AddEntity(a); err != nil {
		log.Fatal(err)
	}
}

func (em *EntityManager) CreateCube() {}

package main

import (
	"log"
	"time"

	"github.com/der-antikeks/gisp/ecs"
)

const (
	GameStateType ecs.ComponentType = iota
)

type GameStateComponent struct {
	State string
	Since time.Time
}

func (c GameStateComponent) Type() ecs.ComponentType {
	return GameStateType
}

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

package storage

import (
	"github.com/gocraft/dbr"
	"github.com/sirupsen/logrus"

	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/events"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/driver/memory"
	postgres "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/driver/postsql"
	eventstorage "github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/driver/postsql/events"
	"github.com/kyma-project/control-plane/components/kyma-environment-broker/internal/storage/postsql"
)

type BrokerStorage interface {
	Instances() Instances
	Operations() Operations
	Provisioning() Provisioning
	Deprovisioning() Deprovisioning
	Orchestrations() Orchestrations
	RuntimeStates() RuntimeStates
	Events() Events
}

const (
	connectionRetries = 10
)

func NewFromConfig(cfg Config, evcfg events.Config, cipher postgres.Cipher, log logrus.FieldLogger) (BrokerStorage, *dbr.Connection, error) {
	log.Infof("Setting DB connection pool params: connectionMaxLifetime=%s "+
		"maxIdleConnections=%d maxOpenConnections=%d", cfg.ConnMaxLifetime, cfg.MaxIdleConns, cfg.MaxOpenConns)

	connection, err := postsql.InitializeDatabase(cfg.ConnectionURL(), connectionRetries, log)
	if err != nil {
		return nil, nil, err
	}

	connection.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	connection.SetMaxIdleConns(cfg.MaxIdleConns)
	connection.SetMaxOpenConns(cfg.MaxOpenConns)

	fact := postsql.NewFactory(connection)

	operation := postgres.NewOperation(fact, cipher)
	return storage{
		instance:       postgres.NewInstance(fact, operation, cipher),
		operation:      operation,
		orchestrations: postgres.NewOrchestrations(fact),
		runtimeStates:  postgres.NewRuntimeStates(fact, cipher),
		events:         events.New(evcfg, eventstorage.New(fact, log)),
	}, connection, nil
}

func NewMemoryStorage() BrokerStorage {
	op := memory.NewOperation()
	return storage{
		operation:      op,
		instance:       memory.NewInstance(op),
		orchestrations: memory.NewOrchestrations(),
		runtimeStates:  memory.NewRuntimeStates(),
	}
}

type storage struct {
	instance       Instances
	operation      Operations
	orchestrations Orchestrations
	runtimeStates  RuntimeStates
	events         Events
}

func (s storage) Instances() Instances {
	return s.instance
}

func (s storage) Operations() Operations {
	return s.operation
}

func (s storage) Provisioning() Provisioning {
	return s.operation
}

func (s storage) Deprovisioning() Deprovisioning {
	return s.operation
}

func (s storage) Orchestrations() Orchestrations {
	return s.orchestrations
}

func (s storage) RuntimeStates() RuntimeStates {
	return s.runtimeStates
}

func (s storage) Events() Events {
	return s.events
}

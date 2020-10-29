package actor

import (
	"context"
	"log"
	"testing"
)

func TestActor_ActorPlusOne(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	// 0 + 1
	out, err := intPlusOne.Call(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	outInt, ok := out.(int)
	if !ok {
		t.Fatal("error cast out")
	}
	if outInt != 1 {
		t.Fatalf("expected: %d actual: %d", 1, outInt)
	}

	// 1 + 1
	out, err = intPlusOne.Call(nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	outInt, ok = out.(int)
	if !ok {
		t.Fatal("error cast out")
	}
	if outInt != 2 {
		t.Fatalf("expected: %d actual: %d", 2, outInt)
	}
}

func TestConnectorActorsPlusOne(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwo := NewActorsConnector(intPlusOne, intPlusOne)
	intPlusThree := NewActorsConnector(intPlusTwo, intPlusOne)
	intPlusSix := NewActorsConnector(intPlusThree, intPlusThree)

	// 0 + 2
	out, err := intPlusTwo.Call(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	outInt, ok := out.(int)
	if !ok {
		t.Fatal("error cast out")
	}
	if outInt != 2 {
		t.Fatalf("expected: %d actual: %d", 2, outInt)
	}

	// 0 + 3
	out, err = intPlusThree.Call(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	outInt, ok = out.(int)
	if !ok {
		t.Fatal("error cast out")
	}
	if outInt != 3 {
		t.Fatalf("expected: %d actual: %d", 3, outInt)
	}

	// 0 + 6
	out, err = intPlusSix.Call(nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	outInt, ok = out.(int)
	if !ok {
		t.Fatal("error cast out")
	}
	if outInt != 6 {
		t.Fatalf("expected: %d actual: %d", 6, outInt)
	}
}

func TestDaemon(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusOneDaemon, err := intPlusOne.AsActorFn().AsDaemon().Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	select {
	case intPlusOneDaemon.In() <- 0:
	}

	select {
	case err := <-intPlusOneDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}
	case out, ok := <-intPlusOneDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 1 {
			t.Fatalf("expected: %d actual: %d", 1, outInt)
		}
	}

	intPlusOneDaemon.Stop()
	intPlusOneDaemon.Wait()

	intPlusOneDaemon, err = intPlusOneDaemon.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	select {
	case intPlusOneDaemon.In() <- 0:
	}

	select {
	case err := <-intPlusOneDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}
	case out, ok := <-intPlusOneDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 1 {
			t.Fatalf("expected: %d actual: %d", 1, outInt)
		}
	}

	intPlusOneDaemon.Stop()
	intPlusOneDaemon.Wait()
}

func TestConnectorDaemonsPlusOne(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwoDaemon, err := NewDaemonsConnector(
		intPlusOne.AsActorFn().AsDaemon(),
		intPlusOne.AsActorFn().AsDaemon(),
	).Run(context.Background())
	if err != nil {
		t.Fatalf("error launch daemon: %s", err)
	}

	intPlusTwoDaemon.In() <- 0

	select {
	case err := <-intPlusTwoDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusTwoDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 2 {
			t.Fatalf("expected: %d actual: %d", 2, outInt)
		}
	}

	intPlusTwoDaemon.In() <- 2

	select {
	case err := <-intPlusTwoDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusTwoDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 4 {
			t.Fatalf("expected: %d actual: %d", 2, outInt)
		}
	}

	intPlusTwoDaemon.Stop()
	intPlusTwoDaemon.Wait()

	////////////////////////////////////////

	intPlusTwo1 := NewDaemonsConnector(
		intPlusOne.AsActorFn().AsDaemon(),
		intPlusOne.AsActorFn().AsDaemon(),
	)

	intPlusFour := NewDaemonsConnector(intPlusTwo1, intPlusTwo1)
	intPlusEight := NewDaemonsConnector(intPlusFour, intPlusFour)

	intPlusFourDaemon, err := intPlusEight.Run(context.Background())
	if err != nil {
		t.Fatalf("error launch daemon: %s", err)
	}

	select {
	case intPlusFourDaemon.In() <- 0:
	}

	select {
	case err := <-intPlusFourDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusFourDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 8 {
			t.Fatalf("expected: %d actual: %d", 8, outInt)
		}

	}

	////////////////////////////////////////

	intPlusTwo := NewDaemonsConnector(
		intPlusOne.AsActorFn().AsDaemon(),
		intPlusOne.AsActorFn().AsDaemon(),
	)
	intPlusFour = NewDaemonsConnector(intPlusTwo, intPlusTwo)
	intPlusEight = NewDaemonsConnector(intPlusFour, intPlusFour)

	intPlusFourSixteen, err := NewDaemonsConnector(intPlusEight, intPlusEight).Run(context.Background())
	if err != nil {
		t.Fatalf("error launch daemon: %s", err)
	}

	intPlusFourSixteen.In() <- 0

	select {
	case err := <-intPlusFourSixteen.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusFourSixteen.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 16 {
			t.Fatalf("expected: %d actual: %d", 16, outInt)
		}

	}

	intPlusFourSixteen.Stop()
	intPlusFourSixteen.Wait()

}

func TestConnectorActorDaemonPlusOne(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwoDaemon, err := NewActorDaemonConnector(
		intPlusOne,
		intPlusOne.AsActorFn().AsDaemon(),
	).Run(context.Background())
	if err != nil {
		log.Fatal("error launch daemon")
	}

	// first try
	intPlusTwoDaemon.In() <- 0
	select {
	case err := <-intPlusTwoDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusTwoDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 2 {
			t.Fatalf("expected: %d actual: %d", 2, outInt)
		}
	}

	//second try:
	intPlusTwoDaemon.In() <- 0

	select {
	case err := <-intPlusTwoDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusTwoDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 2 {
			t.Fatalf("expected: %d actual: %d", 2, outInt)
		}
	}

	intPlusTwoDaemon.Stop()
	intPlusTwoDaemon.Wait()
}

func TestConnectorDaemonActorPlusOne(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwoDaemon, err := NewDaemonActorConnector(
		intPlusOne.AsActorFn().AsDaemon(),
		intPlusOne,
	).Run(context.Background())
	if err != nil {
		log.Fatal("error launch daemon")
	}

	intPlusTwoDaemon.In() <- 0
	select {
	case err := <-intPlusTwoDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusTwoDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 2 {
			t.Fatalf("expected: %d actual: %d", 2, outInt)
		}

	}

	intPlusTwoDaemon.In() <- 5
	select {
	case err := <-intPlusTwoDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusTwoDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 7 {
			t.Fatalf("expected: %d actual: %d", 7, outInt)
		}
	}

	intPlusTwoDaemon.Stop()
	intPlusTwoDaemon.Wait()
}

func TestConnectorActorDaemonDaemonActor(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwoActorDaemon := NewActorDaemonConnector(intPlusOne, intPlusOne.AsActorFn().AsDaemon())
	intPlusTwoDaemonActor := NewDaemonActorConnector(intPlusOne.AsActorFn().AsDaemon(), intPlusOne)

	intPlusFour := NewDaemonsConnector(
		intPlusTwoActorDaemon,
		intPlusTwoDaemonActor,
	)

	intPlusFourDaemon, err := intPlusFour.Run(context.Background())
	if err != nil {
		log.Fatal("error launch daemon")
	}

	intPlusFourDaemon.In() <- 0
	select {
	case err := <-intPlusFourDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusFourDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 4 {
			t.Fatalf("expected: %d actual: %d", 4, outInt)
		}

	}

	intPlusFourDaemon.Stop()
	intPlusFourDaemon.Wait()
}

func TestConnectorDaemonActorActorDaemon(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwoDaemonActor := NewDaemonActorConnector(intPlusOne.AsActorFn().AsDaemon(), intPlusOne)
	intPlusTwoActorDaemon := NewActorDaemonConnector(intPlusOne, intPlusOne.AsActorFn().AsDaemon())

	intPlusFour := NewDaemonsConnector(
		intPlusTwoDaemonActor,
		intPlusTwoActorDaemon,
	)

	intPlusFourDaemon, err := intPlusFour.Run(context.Background())
	if err != nil {
		log.Fatal("error launch daemon")
	}

	intPlusFourDaemon.In() <- 0
	select {
	case err := <-intPlusFourDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusFourDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 4 {
			t.Fatalf("expected: %d actual: %d", 4, outInt)
		}

	}

	intPlusFourDaemon.Stop()
	intPlusFourDaemon.Wait()
}

func TestConnectorDaemonActorActorActor(t *testing.T) {
	intPlusOne := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		inInt, ok := in.(int)
		if !ok {
			return nil, ErrorInputFormat
		}

		return inInt + 1, nil
	})

	intPlusTwoDaemonActor := NewDaemonActorConnector(intPlusOne.AsActorFn().AsDaemon(), intPlusOne)
	intPlusTwoActorActor := NewActorsConnector(intPlusOne, intPlusOne)

	intPlusFour := NewDaemonActorConnector(
		intPlusTwoDaemonActor,
		intPlusTwoActorActor,
	)

	intPlusFourDaemon, err := intPlusFour.Run(context.Background())
	if err != nil {
		log.Fatal("error launch daemon")
	}

	intPlusFourDaemon.In() <- 0
	select {
	case err := <-intPlusFourDaemon.Err():
		if err != nil {
			t.Fatal(err)
		}

	case out, ok := <-intPlusFourDaemon.Out():
		if !ok {
			return
		}
		outInt, ok := out.(int)
		if !ok {
			t.Fatal("error cast out")
		} else if outInt != 4 {
			t.Fatalf("expected: %d actual: %d", 4, outInt)
		}

	}

	intPlusFourDaemon.Stop()
	intPlusFourDaemon.Wait()
}

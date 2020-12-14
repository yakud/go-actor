package actor

import (
	"context"
	"fmt"
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

	intPlusTwo := intPlusOne.ConnectActor(intPlusOne)
	intPlusThree := intPlusTwo.ConnectActor(intPlusOne)
	intPlusSix := intPlusTwo.ConnectActor(intPlusTwo).ConnectActor(intPlusTwo)

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

	intPlusOneAsDaemon := intPlusOne.AsActorFn().AsDaemon()
	intPlusOneDaemon, err := intPlusOneAsDaemon.Run(context.Background())
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

	intPlusOneDaemon, err = intPlusOneAsDaemon.Run(context.Background())
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

	fmt.Println("start first try")
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
	fmt.Println("end first try ok")

	fmt.Println("start second try")
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

	fmt.Println("end second try ok")

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

func TestDaemonsConnectorWait(t *testing.T) {
	b := 0

	d, err := NewDaemonsConnector(
		NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
			out <- 1
			out <- 2
			out <- 3
			return nil
		}),

		NewDaemonsConnector(
			NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
				for _b := range in {
					out <- _b.(int) + 1
				}
				return nil
			}),

			NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
				for _b := range in {
					b = _b.(int)
				}
				return nil
			}),
		),
	).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	d.Wait()
	if b != 4 {
		t.Fatal("b is not 4")
	}
}

func TestDaemonsConnectorWithActorFnWait(t *testing.T) {
	b := 0

	d, err := NewDaemonsConnector(
		NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
			out <- 1
			out <- 2
			out <- 3

			return nil
		}),

		NewDaemonsConnector(
			NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
				return in.(int) + 1, nil
			}).AsActorFn().AsDaemon(),

			NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
				for _b := range in {
					b = _b.(int)
				}
				return nil
			}),
		),
	).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	d.Wait()
	if b != 4 {
		t.Fatal("b is not 4")
	}
}
func TestDaemonsClusterSimple(t *testing.T) {
	d, err := NewDaemonsCluster(
		5,
		NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
			return in.(int) + 1, nil
		}).AsActorFn().AsDaemon(),
	).Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	d.In() <- 1
	a, ok := <-d.Out()
	if !ok {
		t.Fatal("ERRRRRR")
	}

	if a != 2 {
		t.Fatal("a is not 2")
	}

	d.In() <- 2
	a, ok = <-d.Out()
	if !ok {
		t.Fatal("ERRRRRR")
	}
	if a != 3 {
		t.Fatal("a is not 3")
	}

	d.Stop()
	d.Wait()
}

func TestDaemonsMultipleCluster(t *testing.T) {
	intPlusOneDaemon := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		return in.(int) + 1, nil
	}).AsActorFn().AsDaemon()

	d, err := NewDaemonsCluster(5, intPlusOneDaemon).
		ConnectDaemon(NewDaemonsCluster(5, intPlusOneDaemon)).
		Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}

	d.In() <- 1
	a, ok := <-d.Out()
	if !ok {
		t.Fatal("ERRRRRR")
	}

	if a != 3 {
		t.Fatal("a is not 3")
	}

	d.Stop()
	d.Wait()
}

func TestDaemonsClusterWithGenerator(t *testing.T) {
	d, err := NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
		out <- 1
		out <- 2
		out <- 3
		return nil
	}).ConnectDaemon(
		NewDaemonsCluster(
			5,
			NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
				return in.(int) + 1, nil
			}).AsActorFn().AsDaemon(),
		),
	).Run(context.Background())

	if err != nil {
		t.Fatal(err)
	}

	var maxA = 0
	for a := range d.Out() {
		if maxA < a.(int) {
			maxA = a.(int)
		}
	}
	if maxA != 4 {
		t.Fatal("a is not 4")
	}

	d.Wait()
}

func TestBroadcast(t *testing.T) {
	intPlusOneDaemon := NewActor(func(ctx context.Context, in interface{}) (out interface{}, err error) {
		return in.(int) + 1, nil
	}).AsActorFn().AsDaemon()

	adderChan := make(chan interface{})
	intPlusOneDaemon.SetOut(adderChan).DisableCloseChannelsOnStop(true)

	adderDaemon, _ := NewDaemon(func(ctx context.Context, in chan interface{}, out chan interface{}, err chan error) error {
		var sumState = 0

		for inData := range in {
			sumState += inData.(int)
			out <- sumState
		}

		return nil
	}).SetIn(adderChan).Run(context.Background())

	broadcast, _ := NewBroadcastDaemon(
		intPlusOneDaemon.ConnectDaemon(adderDaemon),
		intPlusOneDaemon.ConnectDaemon(adderDaemon),
		intPlusOneDaemon.ConnectDaemon(adderDaemon),
	).Run(context.Background())

	broadcast.In() <- 1

	lastState := <-adderDaemon.Out()
	lastState = <-adderDaemon.Out()
	lastState = <-adderDaemon.Out()

	if lastState != 6 {
		t.Fatal("last state is not 4")
	}

	adderDaemon.Close()
	adderDaemon.Stop()
	adderDaemon.Wait()

	broadcast.Stop()
	broadcast.Wait()
}

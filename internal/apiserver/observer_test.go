package apiserver

import "testing"

func Test_observer(t *testing.T) {

	testObserver1 := &TestObserver{1, nil}
	testObserver2 := &TestObserver{2, nil}
	testObserver3 := &TestObserver{3, nil}
	publisher := Publisher{}

	t.Run("AddSubscriber", func(t *testing.T) {
		publisher.AddSubscriber(testObserver1)
		publisher.AddSubscriber(testObserver2)
		publisher.AddSubscriber(testObserver3)

		if len(publisher.subs) != 3 {
			t.Fail()
		}
	})

	t.Run("RemoveObserver", func(t *testing.T) {
		publisher.RemoveObserver(testObserver2)

		if len(publisher.subs) != 2 {
			t.Fail()
		}
	})

	t.Run("Notify", func(t *testing.T) {
		for _, observer := range publisher.subs {
			printObserver, _ := observer.(*TestObserver)
			message := "hello"
			publisher.NotifyObservers(BusEvent{data: []byte(message)})

			actMessage := string(printObserver.Message)

			if printObserver.Message == nil {
				t.Error()
			}

			if actMessage != message {
				t.Error()
			}
		}
	})
}

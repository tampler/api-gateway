package apiserver

import "testing"

func Test_event(t *testing.T) {

	testObserver1 := &TestObserver{1, ""}
	testObserver2 := &TestObserver{2, ""}
	testObserver3 := &TestObserver{3, ""}
	publisher := Publisher{}

	t.Run("AddSubscriber", func(t *testing.T) {
		publisher.AddSubscriber(testObserver1)
		publisher.AddSubscriber(testObserver2)
		publisher.AddSubscriber(testObserver3)

		if len(publisher.ObserverList) != 3 {
			t.Fail()
		}
	})

	t.Run("RemoveObserver", func(t *testing.T) {
		publisher.RemoveObserver(testObserver2)

		if len(publisher.ObserverList) != 2 {
			t.Fail()
		}
	})

	t.Run("Notify", func(t *testing.T) {
		for _, observer := range publisher.ObserverList {
			printObserver, _ := observer.(*TestObserver)
			message := "hello"
			publisher.NotifyObservers(message)

			if printObserver.Message == "" {
				t.Error()
			}

			if printObserver.Message != message {
				t.Error()
			}
		}
	})

}

package main

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/BradErz/test-mongo/database"
	"github.com/sirupsen/logrus"
)

func main() {
	// create some users to add to the database and peform the test on
	var users []database.User
	for i := 1; i <= 10000; i++ {
		users = append(users, database.NewUser(i))
	}
	logrus.Infof("created %d users", len(users))
	mongoDbUri := "mongodb://mongoadmin:secret@172.17.0.1:27017/test-mongo?authSource=admin"
	if os.Getenv("MONGODB_URI") != "" {
		mongoDbUri = os.Getenv("MONGODB_URI")
	}
	db, err := database.NewDatabase(mongoDbUri)
	if err != nil {
		logrus.WithError(err).Fatal("error connecting to database")
	}
	// remove previously added users
	logrus.Info("dropping previous users in database")
	if err := db.Drop(context.Background()); err != nil {
		logrus.WithError(err).Fatal("error deleting previous database")
	}

	// add all of our test data to mongo
	logrus.Infof("adding %d users to db", len(users))
	for _, u := range users {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
		if err := db.AddUser(ctx, u); err != nil {
			logrus.WithError(err).Fatal("error adding user to database")
			cancel()
		}
		cancel()
	}

	// Create a set of workers to cause simultaneous load on the database
	var wg sync.WaitGroup
	jobCount := 10000
	workerCount := 10

	type WorkerUnit struct {
		Id string
	}

	type Work struct {
		WorkerUnit WorkerUnit
	}

	workers := make(chan Work)

	// start 5 workers to get users
	for x := 1; x <= workerCount; x++ {
		wg.Add(1)
		go func() {
			for w := range workers {
				// create a context with a timeout of 1 second to prove that the query takes over 1 second to complete
				// this is where the error is caused:
				// auth error: sasl conversation error: unable to authenticate using mechanism \"SCRAM-SHA-1\": context deadline exceeded"
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				user, err := db.GetUser(ctx, w.WorkerUnit.Id)
				if err != nil {
					logrus.WithError(err).Error("error getting users from database")
				} else {
					logrus.Infof("found %s users from query", user.Nickname)
				}
				cancel()
			}

		}()
		wg.Done()
	}

	// loop over all the users we created and add the work to the pool for the workers to pick up
	for j := 1; j <= jobCount; j++ {
		for _, u := range users {
			workers <- Work{WorkerUnit: WorkerUnit{Id: u.Id}}
		}
	}

	close(workers)
	wg.Wait()

	logrus.Info("finished")
}

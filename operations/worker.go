package operations

import (
	"strings"
	"time"

	"github.com/mongodb/amboy"
	"github.com/pkg/errors"
	"github.com/tychoish/grip"
	"github.com/tychoish/sink"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
)

func Worker() cli.Command {
	return cli.Command{
		Name: "worker",
		Usage: strings.Join([]string{
			"run a data processing node without a web front-end",
			"runs jobs until there is no more pending work, or 1 minute, whichever is longer",
		}, "\n"),
		Flags: baseFlags(),
		Action: func(c *cli.Context) error {
			ctx := context.Background()
			workers := c.Int("workers")
			mongodbURI := c.String("dbUri")
			bucket := c.String("bucket")

			if err := configure(workers, false, mongodbURI, bucket); err != nil {
				return errors.WithStack(err)
			}

			q, err := sink.GetQueue()
			if err != nil {
				return errors.Wrap(err, "problem getting queue")
			}

			if err := q.Start(ctx); err != nil {
				return errors.Wrap(err, "problem starting queue")
			}

			time.Sleep(time.Minute)
			grip.Debugf("%+v", q.Stats())
			amboy.WaitCtxInterval(ctx, q, time.Second)
			grip.Notice("no pending work; shutting worker down.")

			return nil
		},
	}
}
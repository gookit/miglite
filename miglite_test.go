package miglite_test

import (
	"github.com/gookit/goutil"
	"github.com/gookit/miglite"
	"github.com/gookit/miglite/pkg/command"
	"github.com/gookit/miglite/pkg/config"
)

func ExampleNew() {
	mig, err := miglite.New("miglite.yml", func(cfg *config.Config) {
		// update config options
	})
	goutil.PanicIfErr(err) // handle error

	// run up migrations
	err = mig.Up(command.UpOption{
		Yes: true, // dont confirm
		// ... options
	})
	goutil.PanicIfErr(err) // handle error

	// other operations


	// init migrations schema
	err = mig.Init(command.InitOption{
		// ... options
	})
	goutil.PanicIfErr(err) // handle error

	// run down migrations
	err = mig.Down(command.DownOption{
		// ... options
	})
	goutil.PanicIfErr(err)

	err = mig.Status(command.StatusOption{
		// ... options
	})
	goutil.PanicIfErr(err) // handle error

	err = mig.Show(command.ShowOption{
		// ... options
	})
	goutil.PanicIfErr(err) // handle error

	err = mig.Skip(command.SkipOption{
		// ... options
	})
	goutil.PanicIfErr(err) // handle error
}

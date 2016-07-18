package commands

import (
	"os"
	"sort"
	"time"

	"gopkg.in/cheggaaa/pb.v1"

	"github.com/QuentinPerez/c14-cli/pkg/api"
	"github.com/apex/log"
)

type freeze struct {
	Base
	freezeFlags
}

type freezeFlags struct {
	flQuiet  bool
	flNoWait bool
}

// Freeze returns a new command "freeze"
func Freeze() Command {
	ret := &freeze{}
	ret.Init(Config{
		UsageLine:   "freeze [OPTIONS] [ARCHIVE]+",
		Description: "",
		Help:        "",
		Examples: `
        $ c14 freeze 83b93179-32e0-11e6-be10-10604b9b0ad9`,
	})
	ret.Flags.BoolVar(&ret.flQuiet, []string{"q", "-quiet"}, false, "")
	ret.Flags.BoolVar(&ret.flNoWait, []string{"-nowait"}, false, "")
	return ret
}

func (f *freeze) GetName() string {
	return "freeze"
}

func (f *freeze) CheckFlags(args []string) (err error) {
	if len(args) == 0 {
		f.PrintUsage()
		os.Exit(1)
	}
	return
}

func (f *freeze) Run(args []string) (err error) {
	if err = f.InitAPI(); err != nil {
		return
	}

	var (
		safe        api.OnlineGetSafe
		archiveWait api.OnlineGetArchive
		archives    []api.OnlineGetArchive
		uuidArchive string
	)

	for _, archive := range args {
		if safe, uuidArchive, err = f.OnlineAPI.FindSafeUUIDFromArchive(archive, true); err != nil {
			if safe, uuidArchive, err = f.OnlineAPI.FindSafeUUIDFromArchive(archive, false); err != nil {
				return
			}
		}
		diff := time.Now()
		now := time.Now()
		if err = f.OnlineAPI.PostArchive(safe.UUIDRef, uuidArchive); err != nil {
			log.Warnf("%s: %s", args, err)
			continue
		}
		for now.Before(diff) {
			if archives, err = f.OnlineAPI.GetArchives(safe.UUIDRef, false); err != nil {
				log.Warnf("%s: %s", args, err)
				continue
			}
			sort.Sort(api.OnlineGetArchives(archives))
			uuidArchive = archives[0].UUIDRef
			diff, _ = time.Parse(time.RFC3339, archives[0].CreationDate)
		}

		if !f.flNoWait {
			var bar *pb.ProgressBar

			if !f.flQuiet {
				bar = pb.New(100).SetWidth(80).SetMaxWidth(80).Format("[=> ]")
				bar.ShowFinalTime = false
				bar.ShowTimeLeft = false
				bar.Start()
			}
			lastLength := 6
			for {
				if archiveWait, err = f.OnlineAPI.GetArchive(safe.UUIDRef, uuidArchive, false); err != nil {
					log.Warnf("%s: %s", args, err)
					err = nil
					break
				}
				if lastLength != len(archiveWait.Jobs) {
					lastLength = len(archiveWait.Jobs)
					if !f.flQuiet {
						bar.Add(20)
					}
					if len(archiveWait.Jobs) == 0 {
						break
					}
				}
				time.Sleep(1 * time.Second)
			}
			if !f.flQuiet {
				bar.Finish()
			}
		}
	}
	return
}

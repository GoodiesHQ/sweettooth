package main

import (
	"github.com/goodieshq/sweettooth/pkg/api/client"
	"github.com/goodieshq/sweettooth/pkg/choco"
	"github.com/goodieshq/sweettooth/pkg/schedule"
	"github.com/goodieshq/sweettooth/pkg/util"
	"github.com/rs/zerolog/log"
)

func doPackageJobs(cli *client.SweetToothClient) error {
	log.Trace().Str("routine", "doPackageJobs").Msg("called")
	defer log.Trace().Str("routine", "doPackageJobs").Msg("finished")

	joblist, err := cli.GetPackageJobs()
	if err != nil {
		log.Error().Err(err).Msg("failed to get package jobs")
		return err
	}
	log.Debug().RawJSON("jobs", []byte(util.Dumps(joblist))).Msg("got job list from server")

	// TODO: actually perform the jobs
	for _, jobid := range joblist {
		if !schedule.Now() /* TODO: delete this: */ && false {
			log.Warn().Msg("scheduled maintenance has ended, no longer processing jobs")
			break
		}
		log := log.With().Str("jobid", jobid.String()).Logger()
		log.Trace().Msg("getting job details")

		job, err := cli.GetPackageJob(jobid)
		if err != nil {
			log.Error().Err(err).Msg("failed to get package job")
			return err
		}

		result := choco.Package(choco.PkgAction(job.Action), &job.Parameters)
		cli.CompletePackageJob(jobid, result)

	}

	// choco.Package()

	return nil
}

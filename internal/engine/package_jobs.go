package engine

import (
	"context"

	"github.com/goodieshq/sweettooth/internal/choco"
	"github.com/goodieshq/sweettooth/internal/schedule"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
)

func (engine *SweetToothEngine) PackageJobs() error {
	log.Trace().Str("routine", "PackageJobs").Msg("called")
	defer log.Trace().Str("routine", "PackageJobs").Msg("finished")

	joblist, err := engine.client.GetPackageJobs()
	if err != nil {
		log.Error().Err(err).Msg("failed to get package jobs")
		return err
	}
	log.Debug().RawJSON("jobs", []byte(util.Dumps(joblist))).Msg("got job list from server")

	// TODO: actually perform the jobs
	for _, jobid := range joblist {
		engine.mustRun()

		if !schedule.Now() /* TODO: delete this: */ && false {
			log.Warn().Msg("scheduled maintenance has ended, no longer processing jobs")
			break
		}
		log := log.With().Str("jobid", jobid.String()).Logger()
		log.Trace().Msg("getting job details")

		job, err := engine.client.GetPackageJob(jobid)
		if err != nil {
			log.Error().Err(err).Msg("failed to get package job")
			return err
		}

		// if the code reaches this point, the server has counted it as an attempt

		// TODO figure out a way to safely cancel the chocolatey job here...
		result := choco.Package(context.Background(), choco.PkgAction(job.Action), &job.Parameters)
		engine.client.CompletePackageJob(jobid, result)
	}

	// choco.Package()

	return nil
}

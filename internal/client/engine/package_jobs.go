package engine

import (
	"time"

	"github.com/goodieshq/sweettooth/internal/client/choco"
	"github.com/goodieshq/sweettooth/internal/client/schedule"
	"github.com/goodieshq/sweettooth/internal/util"
	"github.com/rs/zerolog/log"
)

const (
	TIMEOUT_BUFFER_JOB = 30 * time.Second // give an added buffer
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

	for _, jobid := range joblist {
		engine.mustRun()

		if !schedule.Matches() && !BYPASS_SCHEDULE {
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

		// run the package job
		ctx, cancel := engine.commandContext("choco", TIMEOUT_BUFFER_JOB+(time.Second*time.Duration(job.Parameters.Timeout)))
		defer cancel()

		result := choco.Package(
			ctx,
			choco.PkgAction(job.Action),
			&job.Parameters,
		)

		engine.client.CompletePackageJob(jobid, result)
	}

	// choco.Package()

	return nil
}

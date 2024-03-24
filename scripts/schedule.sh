#!/bin/bash
# this should be called by driver.sh, which sets environment variables

#SBATCH --output=output_%A_%a.txt
#SBATCH --cpus-per-task=6
#SBATCH --requeue
#SBATCH --time=48:00:00

job=$SLURM_ARRAY_TASK_ID
start=$(( ($job - 1) * $delta ))
stop=$(( $job * $delta ))

if [ $job -eq $tasks ]
then
    stop=$maxwork
fi

./comma -s $start -t $stop $b

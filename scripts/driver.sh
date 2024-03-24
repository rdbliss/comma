#!/bin/bash
# submit a job array of many path processors (each of which contains many "real" path processors)

# number of parallel processes to spawn
export tasks=300
# base to confirm
export b=23

export maxwork=$(./comma -p $b)
export delta=$(( $maxwork / $tasks ))

if [ ! -d ./results/$b ]; then
    mkdir ./results/$b
fi

# -e/-o redirect stderr and stdout, respectively
sbatch --job-name="comma-$b" --array=1-$tasks --mem=500 -e "./results/$b/error-%a.txt" -o "./results/$b/res-%a.txt" schedule.sh

# cat "./results/$b/res"* | awk '{sum += $1; total = $2} END {print sum,total;}' > "./results/$b/total.txt"

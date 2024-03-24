This directory contains two scripts designed to run the comma sequence prover
on computing clusters which use the [Slurm workload
manager](https://slurm.schedmd.com/documentation.html). Their goal is to spawn
many parallel processes via an sbatch job array to process fixed bases. The
workflow is something like this:

1. Place these scripts and the compiled `comma` binary in the same directory.
2. Edit `driver.sh` with "meta-information," including the desired base to
   check, the number of parallel processes to spawn (tasks), the amount of
   memory each task needs, and so on.
3. Run `driver.sh` and wait for jobs to finish.

When the jobs are done, you can compile the results using some shell commands.
Here's a demo:

```
$ ./driver.sh
$ # wait a long time
$ cat "./results/BASE/res"* | awk '{sum += $1; total = $2} END {print sum,total;}'
```

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

```bash
$ vim driver.sh # change b to, say, 13
$ ./driver.sh # now wait a long time
$ cat "./results/13/res"* | awk '{sum += $1; total = $2} END {print sum,total;}'
```

For a real example, this directory also contains the results of our
computations for base 23. To confirm that all comma sequences in base 23 are
finite, try the following:

```bash
$ cat results/23/res* | awk '{sum += $1; total = $2} END {printf "%.0f %.0f\n", sum, total}'
118291701528000 118291701528000
```

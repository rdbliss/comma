This is the repository of the program `comma`, designed to prove that comma
sequences in arbitrary bases are finite. It uses an algorithm suggested by the
great paper [The Comma Sequence: A Simple Sequence With Bizarre
Properties](https://arxiv.org/abs/2401.14346).

The program is implemented in [go](https://go.dev/). Once you have that
installed, `comma` can be run as follows:

```bash
$ cd comma # or wherever you place this repository
$ go get
$ go build comma.go
$ ./comma 10
 100% |█████████████████████████████████████| (7392/7392, 2438306 it/s)        
49896 49896
```

This output, specifically that the two integers at the end are equal, shows
that all comma sequences in base 10 are finite. This output will likely appear
near-instantaneously, but for a larger base 13, the output includes a progress
bar.

```bash
$ ./comma 13
  18% |████                    | (91020747/487567080, 2436646 it/s) [37s:2m42s]
```

`comma` is a multithreaded program, but beyond base 15 or so, it is highly
recommended that you increase the parallelization by running multiple instances
on the same base at once within a large computing cluster. `comma` accepts
a set "work range" to allow this. Here is an example demonstrating the
functionality, but not the parallelization:

```bash
$ ./comma -p 23 # how much work is there for base 23?
9033184480320
$ ./comma -t 903318448032 23 # work on exactly 1 / 10th of the interval
$ ./comma -s 903318448032 -t 1806636896064 23 # work the next 1 / 10th
```

The `scripts` directory contains the scripts we used to process bases 3 to 23
(excluding 20 and 21) on the Slurm-based [Rutgers Amarel
cluster](https://oarc.rutgers.edu/resources/amarel/).

`comma` accepts a few other options for interested tinkerers. Run `comma
--help` to see a full list.

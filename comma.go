package main

import (
    "runtime"
    "fmt"
    "sync"
    "sync/atomic"
    "time"
    "github.com/schollz/progressbar/v3"
    "math"
    "strconv"
    "runtime/pprof"
    "os"
    "log"
    "github.com/pborman/getopt/v2"
)

type OrderKey struct {
    d int
    uModB int
}

type OrderCached struct {
    ord int
    start int
}

type CycKey struct {
    d int
    unit_digit int
}

type CycCache map[CycKey][]int

type PathParams struct {
    d int
    k int
    b int
    L int
    orderCache [][]OrderCached
    kCache [][][]int
    cyc_cache CycCache
}

type PathPoint struct {
    d int
    u int
    k int
}

type WorkParams struct {
    Workers int
    Feeders int
    Readers int
}

func is_mine(b int, u int) bool {
    reduced_u := ((-u) % (b*b) + (b*b)) % (b*b)
    y := reduced_u % b
    x := reduced_u / b

    return x > 0 && y > 0 && x + y == b - 1
}

// cache all of the transformations (d, u) -> (l, k), where b^i is periodic mod
// S(d, u) with period l for i >= k.
func cacheOrders(b int, cycles CycCache) [][]OrderCached {
    orderCache := make([][]OrderCached, b)
    for d := 1; d < b; d++ {
        orderCache[d] = make([]OrderCached, b)
        for u := 0; u < b; u++ {
            cycle := cycles[CycKey{d, u}]
            S := slicesum(cycle)
            ord, start := order(b, S)
            orderCache[d][u] = OrderCached {ord, start}
        }
    }

    return orderCache
}

// compute the order of b^k mod S, including the initial k value when the
// period kicks in.
func order(b int, S int) (int, int) {
    // find smallest k such that gcd(b^i, S) is constant for i >= k.
    val := 1
    next := b
    k := 0
    for gcd(val, S) < gcd(next, S) {
        val, next = next, (next * b) % S
        k++
    }

    // find order of b mod S / g, where g is the constant gcd.
    S = S / gcd(S, val)
    j := 1
    val = b % S
    for val > 1 {
        // fmt.Println(val)
        val = (val * b) % S
        j++
        if j > S {
            fmt.Println("seem stuck")
            fmt.Println(b, S)
            fmt.Println(val)
            fmt.Println(k)
            time.Sleep(5 * time.Second)
        }
    }

    return j, k
}

func slicesum(slice []int) int {
    sum := 0
    for _, v := range slice {
        sum += v
    }

    return sum
}

func modexp(b int, p int, m int) int {
    res := 1
    val := b
    for p > 0 {
        if p % 2 == 1 {
            res = (res * val) % m
        }
        val = (val * val) % m
        p >>= 1
    }

    return res
}

func process_path(params PathParams) int {
    point := PathPoint {params.d, 0, params.k}

    count := 0
    for {
        // fmt.Println(point)
        if point.d == 1 {
            count++
            if is_mine(params.b, point.u) {
                break
            }
        }

        point = advance_point(point, params)
    }

    return count
}

func process_paths(params chan PathParams, count_reports chan int, wg *sync.WaitGroup) {
    for {
        arg, open := <- params

        if !open {
            (*wg).Done()
            return
        }

        count_reports <- process_path(arg)
    }
    
}

func realMod(x int, y int) int {
    return ((x % y) + y) % y
}

func advance_point(point PathPoint, params PathParams) PathPoint {
    ord := params.orderCache[point.d][point.u % params.b]

    var new_d int
    var new_u int
    var new_k int
    if point.d == params.b - 1 {
        new_d = 1
        new_k = (point.k + 1) % params.L
    } else {
        new_d = point.d + 1
        new_k = point.k % params.L
    }

    if ord.ord == 0 {
        fmt.Println(ord)
        fmt.Println(point)
        // fmt.Println(params)
        time.Sleep(5 * time.Second)
    }
    reduced_k := point.k % ord.ord
    new_u = params.kCache[point.d][point.u][reduced_k]

    return PathPoint{new_d, new_u, new_k}
}

func lcm(a int, b int) int {
    return a * b / gcd(a, b)
}

func gcd(a int, b int) int {
    if a == 0 {
        return b
    } else if b == 0 {
        return a
    }

    return gcd(b, a % b)
}

func get_lcm_cycles(b int) (int, CycCache) {
    cycles := make(CycCache)
    L := 1
    for d := 1; d < b; d++{
        for unit_digit := range b {
            cycle_len := b / gcd(b, d)
            cycle := make([]int, cycle_len)
            for m := range cycle_len {
                // for some god-forsaken reason, Go produces negative
                // remainders. that's why we fix that below.
                cycle[m] = d + b * ((((-unit_digit + m * d) % b) + b) % b)
            }

            cycles[CycKey{d, unit_digit}] = cycle
            ord, _ := order(b, slicesum(cycle))
            L = lcm(L, ord)
        }
    }

    return L, cycles
}

func isFinite(b int, params WorkParams) (int, int) {
    L, cycles := get_lcm_cycles(b)
    orderCache := cacheOrders(b, cycles)
    kCache := cacheAdvances(b, cycles, orderCache)

    work := make(chan PathParams)
    report_counts := make(chan int)

    // put work into a queue
    var workWg sync.WaitGroup
    for i := range params.Feeders {
        workWg.Add(1)
        go func() {
            for d := 2 + (i * (b - 2)) / params.Feeders; d < 2 + ((i + 1) * (b - 2)) / params.Feeders; d++{
            for k := 0; k < L; k++ {
                work <- PathParams{d, k, b, L, orderCache, kCache, cycles}
            }
            }
            workWg.Done()
        }()
    }

    var processWg sync.WaitGroup
    for _ = range params.Workers {
        // start workers
        processWg.Add(1)
        go process_paths(work, report_counts, &processWg)
    }

    // listen for count updates
    var countWg sync.WaitGroup
    expected_work := (b - 2) * L
    bar := progressbar.Default(int64(expected_work))

    var count atomic.Uint64

    for _ = range params.Readers {
    countWg.Add(1)
    go func() {
	    for {
            delta, open := <-report_counts
            if !open {
                break
            }
            count.Add(uint64(delta))
            bar.Add(1)
        }

	    countWg.Done()
    }()
    }

    workWg.Wait()
    close(work)
    processWg.Wait()
    close(report_counts)
    countWg.Wait()

    target := valid_count(b) * L
    return int(count.Load()), target
}

func valid_count(b int) int {
    count := 0
    for u:=1; u < b*b; u++ {
        if u <= b * (((-u % b) + b) % b) + b - 1 {
            count++
        }
    }

    return count
}

// create a map of the transforms (d, u, k mod l(d, u)) -> u'
// this map is read-only after return!
func cacheAdvances(b int, cycCache CycCache, orderCache [][]OrderCached) [][][]int {
    k_cache := make([][][]int, b)

    // this might start a *lot* of goroutines.
    // here goes nothing!
    var wg sync.WaitGroup

    for d := 1; d < b; d++ {
        k_cache[d] = make([][]int, b * b)

        var new_d int
        if d == b - 1 {
            new_d = 1
        } else {
            new_d = d + 1
        }

        wg.Add(1)
        go func() {
        defer wg.Done()
        for u := 0; u < b*b; u++ {
            if u > b * (((-u % b) + b) % b) + b - 1 {
                continue
            }

            // determine the correct modulus for k.
            cyc := cycCache[CycKey {d, u % b}]
            S := slicesum(cyc)
            ord := orderCache[d][u % b]

            k_cache[d][u] = make([]int, ord.ord)

            for k := range(ord.ord) {
                wg.Add(1)
                go func() {
                    defer wg.Done()
                    gap := (modexp(b, ord.start + realMod(k - ord.start, ord.ord), S) + u) % S

                    if gap == 0 {
                        gap = cyc[len(cyc) - 1]
                    } else {
                        for _, x := range cyc {
                            if gap <= x {
                                break
                            }

                            if gap < b*b && new_d == 1 && is_mine(b, gap) {
                                break
                            }
                            gap -= x
                        }
                    }

                    new_u := gap
                    k_cache[d][u][k] = new_u
                }()
            }
        }
        }()
    }

    wg.Wait()
    return k_cache
}

func main() {
    helpFlag := getopt.BoolLong("help", 'h', "display help")
    estimateLimit := getopt.IntLong("estimate", 0, 10, "estimate runtime up to b using runtimes from 2 to value")
    workers := getopt.IntLong("workers", 'w', runtime.NumCPU() * 2, "number of workers to use processesing paths")
    feeders := getopt.IntLong("feeders", 'f', runtime.NumCPU(), "number of threads feeding work to workers")
    readers := getopt.IntLong("readers", 'r', runtime.NumCPU(), "number of threads reading results from workers")
    cpuprofile := getopt.StringLong("cpuprofile", 0, "", "write cpu profile to file")
    getopt.SetParameters("b")

    getopt.Parse()

    if *helpFlag {
        getopt.PrintUsage(os.Stderr)
        return
    }

    if *cpuprofile != "" {
        fmt.Println("cpuprofile:", *cpuprofile)
        f, err := os.Create(*cpuprofile)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    args := getopt.Args()

    if len(args) < 1 {
        fmt.Fprintln(os.Stderr, "Error: missing b")
        getopt.PrintUsage(os.Stderr)
        return
    }

    b, err := strconv.Atoi(args[0])

    if err != nil {
        fmt.Println("trouble parsing b")
        getopt.PrintUsage(os.Stderr)
        return
    }

    params := WorkParams {*workers, *feeders, *readers}

    if !getopt.IsSet("estimate") {
        fmt.Println(isFinite(b, params))
        return
    }

    var lcms []float64
    var times []float64

    for k := 2; k <= *estimateLimit; k++ {
        L, _ := get_lcm_cycles(k)
        lcms = append(lcms, float64(L))
        start := time.Now()
        isFinite(k, params)
        times = append(times, float64(time.Since(start) / time.Microsecond))
    }

    var slope float64 = (math.Log(times[len(times) - 1]) - math.Log(times[1])) / (math.Log(lcms[len(lcms) - 1]) - math.Log(lcms[1]))
    var intercept float64 = math.Log(times[1]) - slope * math.Log(lcms[1])
    fmt.Println("slope:", slope)
    fmt.Println("intercept:", intercept)

    fmt.Println("lcms:", lcms)
    fmt.Println("times (microseconds):", times)

    for k := 2; k <= b; k++ {
        L, _ := get_lcm_cycles(k)
        // get microseconds for time.Duration
        estimate := math.Exp(intercept) * math.Pow(float64(L), slope) * 1000
        fmt.Printf("$%d$ & %d & %s \\\\\n", k, L, time.Duration(estimate))
    }
}

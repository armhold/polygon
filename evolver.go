package polygen

import (
	"image"
	"log"
	"sort"
	"math/rand"
	"time"
	"sync"
)

func SimulateAnnealing(maxGen int, referenceImg image.Image, destFile string, safeImages []*SafeImage) {
	refImgRGBA := ConvertToRGBA(referenceImg)

	w := refImgRGBA.Bounds().Dx()
	h := refImgRGBA.Bounds().Dy()

	mostFit := RandomCandidate(w, h)
	evaluateCandidate(mostFit, refImgRGBA)

	candidates := make([]*Candidate, PopulationCount)
	candidates[0] = mostFit

	startTime := time.Now()

	generationsSinceChange := 0

	for gen := 0; gen < maxGen; gen++ {

		c := make(chan *Candidate, PopulationCount)
		var wg sync.WaitGroup

		processCandidate := func(cand *Candidate) {
			cand.MutateInPlace()
			evaluateCandidate(cand, refImgRGBA)
			c <- cand
			wg.Done()
		}

		wg.Add(PopulationCount - 1)

		for i := 1; i < PopulationCount; i++ {
			go processCandidate(mostFit.CopyOf())
		}

		wg.Wait()

		for i := 1; i < PopulationCount; i++ {
			candidates[i] = <- c
		}

		// after sort, the best will be at [0], worst will be at [len() - 1]
		sort.Sort(ByFitness(candidates))

		if gen % 10 == 0 {
			printStats(candidates, gen, generationsSinceChange, startTime)
		}

		for i := 0; i < len(safeImages); i++ {
			safeImages[i].Update(candidates[i].img)
		}

		currBest := candidates[0]

		if currBest.Fitness < mostFit.Fitness {
			generationsSinceChange = 0
			mostFit = currBest
		} else {
			generationsSinceChange++
		}
	}

	mostFit.DrawAndSave(destFile)
	log.Printf("after %d generations, fitness is: %d, saved to %s", maxGen, mostFit.Fitness, destFile)
}


func evaluateCandidate(c *Candidate, referenceImg *image.RGBA) {
//	diff, err := Compare(referenceImg, c.img)
	diff, err := FastCompare(referenceImg, c.img)

	if err != nil {
		log.Fatalf("error comparing images: %s", err)
	}

	c.Fitness = diff
}

func shufflePopulation(population []*Candidate) {
	for i := range population {
		j := rand.Intn(i + 1)
		population[i], population[j] = population[j], population[i]
	}
}

func printStats(sortedPop []*Candidate, generations, generationsSinceChange int, startTime time.Time) {
	dur := time.Since(startTime)
	best := sortedPop[0].Fitness
	worst := sortedPop[len(sortedPop)-1].Fitness

	log.Printf("dur: %s, gen: %d, since change: %d, best: %d, worst: %d", dur, generations, generationsSinceChange, best, worst)
}

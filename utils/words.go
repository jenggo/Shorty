package utils

import (
	"math/rand/v2"
)

// copy from https://github.com/xyproto/randomstring/blob/main/randomstring.go

var (
	random = rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64()))
	freq   = map[rune]int{
		'e': 21912,
		't': 16587,
		'a': 14810,
		'o': 14003,
		'i': 13318,
		'n': 12666,
		's': 11450,
		'r': 10977,
		'h': 10795,
		'd': 7874,
		'l': 7253,
		'u': 5246,
		'c': 4943,
		'm': 4761,
		'f': 4200,
		'y': 3853,
		'w': 3819,
		'g': 3693,
		'p': 3316,
		'b': 2715,
		'v': 2019,
		'k': 1257,
		'x': 315,
		'q': 205,
		'j': 188,
		'z': 128,
	}
	freqVowel = map[rune]int{
		'e': 21912,
		'a': 14810,
		'o': 14003,
		'i': 13318,
		'u': 5246,
	}
	freqCons = map[rune]int{
		't': 16587,
		'n': 12666,
		's': 11450,
		'r': 10977,
		'h': 10795,
		'd': 7874,
		'l': 7253,
		'c': 4943,
		'm': 4761,
		'f': 4200,
		'y': 3853,
		'w': 3819,
		'g': 3693,
		'p': 3316,
		'b': 2715,
		'v': 2019,
		'k': 1257,
		'x': 315,
		'q': 205,
		'j': 188,
		'z': 128,
	}
	freqsum = func() int {
		n := 0
		for _, v := range freq {
			n += v
		}
		return n
	}()
	freqsumVowel = func() int {
		n := 0
		for _, v := range freqVowel {
			n += v
		}
		return n
	}()
	freqsumCons = func() int {
		n := 0
		for _, v := range freqCons {
			n += v
		}
		return n
	}()
)

func pickLetter() rune {
	target := random.IntN(freqsum)
	selected := 'a'
	n := 0
	for k, v := range freq {
		n += v
		if n >= target {
			selected = k
			break
		}
	}
	return selected
}

func pickVowel() rune {
	target := random.IntN(freqsumVowel)
	selected := 'a'
	n := 0
	for k, v := range freqVowel {
		n += v
		if n >= target {
			selected = k
			break
		}
	}
	return selected
}

func pickCons() rune {
	target := random.IntN(freqsumCons)
	selected := 't'
	n := 0
	for k, v := range freqCons {
		n += v
		if n >= target {
			selected = k
			break
		}
	}
	return selected
}

func HumanFriendlyEnglishString(length int) string {
	vowelOffset := random.IntN(2)
	vowelDistribution := 2
	b := make([]byte, length)
	for i := 0; i < length; i++ {
	again:
		switch {
		case (i+vowelOffset)%vowelDistribution == 0:
			b[i] = byte(pickVowel())
		case random.IntN(100) > 0: // 99 of 100 times
			b[i] = byte(pickCons())
			// Don't repeat
			if i >= 1 && b[i] == b[i-1] {
				// Also use more vowels
				vowelDistribution = 1
				// Then try again
				goto again
			}
		default:
			b[i] = byte(pickLetter())
			// Don't repeat
			if i >= 1 && b[i] == b[i-1] {
				// Also use more vowels
				vowelDistribution = 1
				// Then try again
				goto again
			}
		}

		// Avoid three letters in a row
		if i >= 2 && b[i] == b[i-2] {
			// Then try again
			goto again
		}
	}
	return string(b)
}

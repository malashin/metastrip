package main

import (
	"fmt"
	"os"

	"github.com/malashin/metastrip/png"
)

func main() {
	file := os.Args[1]

	p, err := png.Open(file)
	if p == nil {
		panic(err)
	}
	if err != nil {
		fmt.Println(err)
	}
	defer p.Close()

	fmt.Println(p)

	f, err := os.Create(file + "####.png")
	if err != nil {
		panic(err)
	}

	err = png.WriteSignatureTo(f)
	if err != nil {
		panic(err)
	}
	for _, c := range p.Chunks {
		switch c.Type {
		case "IHDR", "PLTE", "tRNS", "pHYs", "IDAT", "IEND":
			err = c.WriteTo(f)
			if err != nil {
				panic(err)
			}
		}
	}

	// j, err := jpg.Open(file)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// if j == nil {
	// 	panic(err)
	// }
	// defer j.Close()

	// fmt.Println(j)

	// f, err := os.Create(file + "####.jpg")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// for _, c := range j.Chunks {
	// 	switch {
	// 	default:
	// 		err = c.WriteTo(f)
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 	case c.Type == jpg.COM || jpg.APP1 <= c.Type && c.Type <= jpg.APP13 || c.Type == jpg.APP15:
	// 		// skip
	// 	}
	// }

	// fj, err := jpg.Open(file + "####.jpg")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// if fj == nil {
	// 	panic(err)
	// }
	// defer fj.Close()

	// fmt.Println(fj)
}

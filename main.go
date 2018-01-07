package main

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

func readHexText(fname string) (bytes.Buffer, error) {
	var bb bytes.Buffer
	f, err := os.Open(fname)
	if err != nil {
		log.Fatal(err)
		return bb, err
	}
	r1 := bufio.NewReader(f)
	for {
		var hexString string
		for i := 0; i < 2; i++ {
			var oneRune rune
			oneRune, _, err = r1.ReadRune()
			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
				break
			}
			hexString += string(oneRune)
		}
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
			break
		}
		oneByte, _ := hex.DecodeString(hexString)
		bb.Write(oneByte)
	}
	f.Close()
	return bb, err
}

func readText(r io.Reader) (int64, bytes.Buffer) {
	var bb bytes.Buffer
	n, err := bb.ReadFrom(r)
	if err != nil {
		log.Fatal(err)
	}
	return n, bb
}

func isPunct(aChar byte) bool {
	if (aChar >= 0x20) && (aChar < 0x30) {
		return true
	}
	if (aChar > 0x39) && (aChar < 0x41) {
		return true
	}
	if (aChar > 0x5a) && (aChar < 0x61) {
		return true
	}
	if (aChar > 0x7a) && (aChar < 0x7f) {
		return true
	}
	return false
}

func isUpperLetter(aChar byte) bool {
	if (aChar > 0x40) && (aChar < 0x5b) {
		return true
	}
	return false
}

func isLowerLetter(aChar byte) bool {
	if (aChar > 0x60) && (aChar < 0x7b) {
		return true
	}
	return false
}

func isLetter(aChar byte) bool {
	if isLowerLetter(aChar) || isUpperLetter(aChar) {
		return true
	}
	return false
}

type charProb map[string]float32

func readCharFrequencyTable(r io.Reader) charProb {
	dec := json.NewDecoder(r)
	var cp charProb
	if err := dec.Decode(&cp); err != nil {
		log.Fatal("trying to decode frequency table", err)
		return nil
	}
	return cp
}

type probPair []struct {
	Char string
	Prob float32
}

func readFrequencyTable(r io.Reader) map[string]float32 {
	dec := json.NewDecoder(r)
	//err := json.Unmarshal()
	var ppm map[string]float32
	ppm = make(map[string]float32)
	for {
		var pp probPair
		if err := dec.Decode(&pp); err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		for _, value := range pp {
			ppm[value.Char] = value.Prob / 100
		}
		//ppm[pp.Char] = pp.Prob
	}
	return ppm
}

type freqStats struct {
	counter, punct, letter, other int16
	ccm                           map[byte]int16
	cfm                           map[byte]float32
}

func characterClassifier(bb bytes.Buffer) freqStats {
	var fs freqStats
	var counter int16
	var b byte
	//var ccm map[byte]int16
	fs.ccm = make(map[byte]int16)
	for b, err := bb.ReadByte(); err != io.EOF; b, err = bb.ReadByte() {
		counter++
		log.Printf("%c %x %d\n", b, b, b)
		switch {
		case isUpperLetter(b):
			fs.letter++
			fs.ccm[b+0x20]++
		case isLowerLetter(b):
			fs.letter++
			fs.ccm[b]++
		case isPunct(b):
			fs.punct++
		default:
			fs.other++
		}
	}
	fs.cfm = make(map[byte]float32)
	for b = range fs.ccm {
		fs.cfm[b] = float32(fs.ccm[b]) / float32(fs.letter)
	}
	return fs
}

func squaredDifference(cfm map[byte]float32, ppm charProb) float32 {
	var diffsum, targetsum float32
	for k, v := range ppm {
		b := byte(k[0])
		rv := v / 100
		println(k, b, v)
		diffsum += rv * cfm[b]
		targetsum += rv * rv
	}
	targetdiff := diffsum - targetsum
	return targetdiff
}

func main() {
	//bb, _ := readHexText("hexMessage.txt")
	//bblen := bb.Len()
	//fmt.Printf("length:%s \n", hex.EncodeToString(bb.Bytes()))
	f, err := os.Open("English_language_frequency_table_map.json")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	r1 := bufio.NewReader(f)
	ppm := readCharFrequencyTable(r1)
	for key, value := range ppm {
		fmt.Println("Key:", key, "Value:", value)
	}

	f, err = os.Open("sample-text.txt")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer f.Close()
	r1 = bufio.NewReader(f)
	_, bb := readText(r1)

	fs := characterClassifier(bb)

	targetdiff := squaredDifference(fs.cfm, ppm)
	println(targetdiff)
}

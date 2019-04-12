package commons

import (
	"strconv"
	"strings"
)

func SplitToInt(line string) ([]int, error) {
	sp := strings.Split(line, ",")
	intsp := make([]int, len(sp))
	for idx, v := range sp {
		if i, err := strconv.Atoi(v); err != nil {
			return nil, err
		} else {
			intsp[idx] = i
		}
	}
	return intsp, nil
}

func SplitToInt2(line string) (int, int, error) {
	if sp, err := SplitToInt(line); err != nil {
		return 0, 0, err
	} else {
		return sp[0], sp[1], nil
	}
}

func SplitToUint(line string, base int, bitSize int) ([]uint64, error) {
	sp := strings.Split(line, ",")
	intsp := make([]uint64, len(sp))
	for idx, v := range sp {
		if i, err := strconv.ParseUint(v, base, bitSize); err != nil {
			return nil, err
		} else {
			intsp[idx] = i
		}
	}
	return intsp, nil
}

func SplitToUint2(line string, base int, bitSize int) (uint64, uint64, error) {
	if sp, err := SplitToUint(line, base, bitSize); err != nil {
		return 0, 0, err
	} else {
		return sp[0], sp[1], nil
	}
}

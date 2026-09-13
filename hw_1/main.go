package main

import "fmt"

func twoSum(nums []int, target int) []int {

	m := make(map[int]int)

	for i := 0; i < len(nums); i++ {

		value_1 := nums[i]
		value_2 := target - value_1

		val, ok := m[value_2]

		if ok {
			return []int{val, i}
		}

		m[value_1] = i

	}
	return nil

}

func main() {

	fmt.Println(twoSum([]int{2, 11, 10, 3}, 5))

}

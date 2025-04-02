package leetcode

func canReplace(s, t string) bool {
	cnt := 0
	for i := 0; i < len(t); i++ {
		if s[i] != t[i] {
			cnt++
		}

		if cnt > 1 {
			return false
		}
	}
	return cnt == 1
}

func canDeleteInsert(s, t string) bool {
	lt := len(t)
	ls := len(s)
	i := 0
	j := 0
	diff := 0

	for i < lt && j < ls {

		if s[i] == s[j] {
			i++
			j++
		} else {
			if diff > 1 {
				return false
			}
			ls++
		}
	}

	return true
}

func Abs(a int) int {
	if a < 0 {
		return -a
	}

	return a
}

func OneEdit(s, t string) bool {

	lt := len(t)
	ls := len(s)
	if Abs(lt-ls) > 1 {
		return false
	}

	if ls == lt {
		return canReplace(s, t)
	}

	if Abs(lt-ls) == 1 {
		return canDeleteInsert(s, t)
	}

	return true

}

// ассимптотическая сложность O(n + m), где n - длина наибольшей строки
// сложность по памяти О(1)

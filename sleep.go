package main
import (
	// "log"
	"math"
	"math/rand"
	"time"
)
// delay秒間スリープする
func sleep(delay int) {
	// msecに変換する
	msec := delay * 1000
	// スリープ時間に揺らぎを与える
	// ゆらぎの最小値をmsecの常用対数の-10倍、最大値を10倍としてゆらぎを乱数で生成する
	// 乱数の生成にはmath/randパッケージを使用する

	fluctuation := int((100 * rand.Float64() - 10) * math.Log10(float64(msec)))
	// log.Printf("sleep: delay=%d, fluctuation=%d, total=%d\n", msec, fluctuation, msec+fluctuation)
	time.Sleep(time.Duration(msec+fluctuation) * time.Millisecond)
}

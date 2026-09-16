package main
import (
	// "log"
	"math"
	"math/rand"
	"time"
)
// delay秒間スリープする
func sleep(delay float64) {
	// msecに変換する
	msec := int(delay * 1000)
	if msec < 200 {
		msec = 200
	}
	// スリープ時間に揺らぎを与える
	// ゆらぎの最小値をmsecの常用対数の-10倍、最大値を10倍としてゆらぎを乱数で生成する
	// 乱数の生成にはmath/randパッケージを使用する

	fluctuation := int((100 * rand.Float64() - 50) * math.Log10(float64(msec)))
	if msec + fluctuation < 100 {
		fluctuation = 100 - msec
	}
	// log.Printf("sleep: delay=%d, fluctuation=%d, total=%d\n", msec, fluctuation, msec+fluctuation)
	time.Sleep(time.Duration(msec+fluctuation) * time.Millisecond)
}

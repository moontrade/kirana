package logger

import "testing"

func doTestLog() {
	Info()
	Err()
	Info0()
	Info0()
	println()

	func() {
		Info()
		Err()
		Info0()
		println()

		func() {
			Info()
			Err()
			Info0()
			println()

			func() {
				Info()
				Err()
				Info0()
				println()
			}()
		}()

		x := func() {
			Info()
			Err()
			Info0()
			println()
		}
		x()
	}()
}

func doTestLogSlow() {
	getCallerPCSlowSlow()
	println()

	func() {
		getCallerPCSlowSlow()
		println()

		func() {
			getCallerPCSlowSlow()
			println()

			func() {
				getCallerPCSlowSlow()
				println()
			}()
		}()

		x := func() {
			getCallerPCSlowSlow()
			println()
		}
		x()
	}()
}

func TestLog(t *testing.T) {
	//Info()
	doTestLog()
	//doTestLogSlow()
}

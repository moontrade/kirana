package spinlock

import (
	"sync"
	"testing"
)

/*
	Mac M1

Fastest for a single thread, but scales worse than relaxed
atomic.CompareAndSwapUint32
BenchmarkSpinLock_Lock/Spinlock_1_thread
BenchmarkSpinLock_Lock/Spinlock_1_thread-8         	140501175	         8.138 ns/op
BenchmarkSpinLock_Lock/Spinlock_2_threads
BenchmarkSpinLock_Lock/Spinlock_2_threads-8        	37530150	        36.43 ns/op
BenchmarkSpinLock_Lock/Spinlock_4_threads
BenchmarkSpinLock_Lock/Spinlock_4_threads-8        	17327236	        69.52 ns/op
BenchmarkSpinLock_Lock/Spinlock_8_threads
BenchmarkSpinLock_Lock/Spinlock_8_threads-8        	 8597024	       136.5 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_1_thread
BenchmarkSpinLock_Lock/RWSpinlock_1_thread-8       	88241509	        13.49 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_2_threads
BenchmarkSpinLock_Lock/RWSpinlock_2_threads-8      	24819048	        48.26 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_4_threads
BenchmarkSpinLock_Lock/RWSpinlock_4_threads-8      	11941560	       100.5 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_8_threads
BenchmarkSpinLock_Lock/RWSpinlock_8_threads-8      	 5646890	       212.5 ns/op
BenchmarkSpinLock_Lock/sync.Mutex
BenchmarkSpinLock_Lock/sync.Mutex-8                	88138082	        13.49 ns/op
BenchmarkSpinLock_Lock/sync.Mutex_2_threads
BenchmarkSpinLock_Lock/sync.Mutex_2_threads-8      	13211614	        94.01 ns/op
BenchmarkSpinLock_Lock/sync.Mutex_4_threads
BenchmarkSpinLock_Lock/sync.Mutex_4_threads-8      	 4607575	       272.2 ns/op
BenchmarkSpinLock_Lock/sync.Mutex_8_threads
BenchmarkSpinLock_Lock/sync.Mutex_8_threads-8      	 1907782	       631.5 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex
BenchmarkSpinLock_Lock/sync.RWMutex-8              	62983033	        18.50 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex_2_threads
BenchmarkSpinLock_Lock/sync.RWMutex_2_threads-8    	 7609563	       163.6 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex_4_threads
BenchmarkSpinLock_Lock/sync.RWMutex_4_threads-8    	 2532800	       472.1 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex_8_threads
BenchmarkSpinLock_Lock/sync.RWMutex_8_threads-8    	 1218156	       983.8 ns/op

Relaxed CAS and Add scales a better
atomic_internal.Cas
BenchmarkSpinLock_Lock/Spinlock_1_thread
BenchmarkSpinLock_Lock/Spinlock_1_thread-8         	137080747	         8.331 ns/op
BenchmarkSpinLock_Lock/Spinlock_2_threads
BenchmarkSpinLock_Lock/Spinlock_2_threads-8        	46482424	        27.03 ns/op
BenchmarkSpinLock_Lock/Spinlock_4_threads
BenchmarkSpinLock_Lock/Spinlock_4_threads-8        	19519190	        65.61 ns/op
BenchmarkSpinLock_Lock/Spinlock_8_threads
BenchmarkSpinLock_Lock/Spinlock_8_threads-8        	 9223419	       132.2 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_1_thread
BenchmarkSpinLock_Lock/RWSpinlock_1_thread-8       	100000000	        11.78 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_2_threads
BenchmarkSpinLock_Lock/RWSpinlock_2_threads-8      	28893673	        46.22 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_4_threads
BenchmarkSpinLock_Lock/RWSpinlock_4_threads-8      	14021223	        91.82 ns/op
BenchmarkSpinLock_Lock/RWSpinlock_8_threads
BenchmarkSpinLock_Lock/RWSpinlock_8_threads-8      	 6700483	       178.5 ns/op
BenchmarkSpinLock_Lock/sync.Mutex
BenchmarkSpinLock_Lock/sync.Mutex-8                	84994611	        13.49 ns/op
BenchmarkSpinLock_Lock/sync.Mutex_2_threads
BenchmarkSpinLock_Lock/sync.Mutex_2_threads-8      	13763782	        94.20 ns/op
BenchmarkSpinLock_Lock/sync.Mutex_4_threads
BenchmarkSpinLock_Lock/sync.Mutex_4_threads-8      	 4458303	       279.7 ns/op
BenchmarkSpinLock_Lock/sync.Mutex_8_threads
BenchmarkSpinLock_Lock/sync.Mutex_8_threads-8      	 1714070	       703.5 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex
BenchmarkSpinLock_Lock/sync.RWMutex-8              	64678983	        18.50 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex_2_threads
BenchmarkSpinLock_Lock/sync.RWMutex_2_threads-8    	10581184	       170.5 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex_4_threads
BenchmarkSpinLock_Lock/sync.RWMutex_4_threads-8    	 2549134	       464.9 ns/op
BenchmarkSpinLock_Lock/sync.RWMutex_8_threads
BenchmarkSpinLock_Lock/sync.RWMutex_8_threads-8    	 1000000	      1046 ns/op
*/
func BenchmarkSpinLock_Lock(b *testing.B) {
	var (
		rw = false
	)
	b.Run("Spinlock 1 thread", func(b *testing.B) {
		l := new(Mutex)
		for i := 0; i < b.N; i++ {
			l.Lock()
			l.Unlock()
		}
	})

	b.Run("Spinlock 2 threads", func(b *testing.B) {
		var (
			l  = new(Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("Spinlock 4 threads", func(b *testing.B) {
		var (
			l  = new(Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("Spinlock 8 threads", func(b *testing.B) {
		var (
			l  = new(Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})

	if rw {
		b.Run("RWSpinlock 1 thread", func(b *testing.B) {
			l := new(RWMutex)
			for i := 0; i < b.N; i++ {
				l.Lock()
				l.Unlock()
			}
		})

		b.Run("RWSpinlock 2 threads", func(b *testing.B) {
			var (
				l  = new(RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.Lock()
						l.Unlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("RWSpinlock 4 threads", func(b *testing.B) {
			var (
				l  = new(RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.Lock()
						l.Unlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("RWSpinlock 8 threads", func(b *testing.B) {
			var (
				l  = new(RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.Lock()
						l.Unlock()
					}
				}()
			}
			wg.Wait()
		})
	}

	b.Run("sync.Mutex", func(b *testing.B) {
		l := new(sync.Mutex)
		for i := 0; i < b.N; i++ {
			l.Lock()
			l.Unlock()
		}
	})
	b.Run("sync.Mutex 2 threads", func(b *testing.B) {
		var (
			l  = new(sync.Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("sync.Mutex 4 threads", func(b *testing.B) {
		var (
			l  = new(sync.Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("sync.Mutex 8 threads", func(b *testing.B) {
		var (
			l  = new(sync.Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})

	if rw {
		b.Run("sync.RWMutex", func(b *testing.B) {
			var (
				l = new(sync.RWMutex)
			)
			for i := 0; i < b.N; i++ {
				l.Lock()
				l.Unlock()
			}
		})
		b.Run("sync.RWMutex 2 threads", func(b *testing.B) {
			var (
				l  = new(sync.RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.Lock()
						l.Unlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("sync.RWMutex 4 threads", func(b *testing.B) {
			var (
				l  = new(sync.RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.Lock()
						l.Unlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("sync.RWMutex 8 threads", func(b *testing.B) {
			var (
				l  = new(sync.RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.Lock()
						l.Unlock()
					}
				}()
			}
			wg.Wait()
		})
	}
}

/*
BenchmarkSpinLock_ReadLock/Spinlock_1_thread
BenchmarkSpinLock_ReadLock/Spinlock_1_thread-8         	136984424	         8.336 ns/op
BenchmarkSpinLock_ReadLock/Spinlock_2_threads
BenchmarkSpinLock_ReadLock/Spinlock_2_threads-8        	42713492	        26.53 ns/op
BenchmarkSpinLock_ReadLock/Spinlock_4_threads
BenchmarkSpinLock_ReadLock/Spinlock_4_threads-8        	16012720	        78.73 ns/op
BenchmarkSpinLock_ReadLock/Spinlock_8_threads
BenchmarkSpinLock_ReadLock/Spinlock_8_threads-8        	 9726903	       121.3 ns/op
BenchmarkSpinLock_ReadLock/RWSpinlock_1_thread
BenchmarkSpinLock_ReadLock/RWSpinlock_1_thread-8       	95038844	        12.55 ns/op
BenchmarkSpinLock_ReadLock/RWSpinlock_2_threads
BenchmarkSpinLock_ReadLock/RWSpinlock_2_threads-8      	29070823	        43.19 ns/op
BenchmarkSpinLock_ReadLock/RWSpinlock_4_threads
BenchmarkSpinLock_ReadLock/RWSpinlock_4_threads-8      	12296955	       105.6 ns/op
BenchmarkSpinLock_ReadLock/RWSpinlock_8_threads
BenchmarkSpinLock_ReadLock/RWSpinlock_8_threads-8      	 2776194	       430.7 ns/op
BenchmarkSpinLock_ReadLock/sync.Mutex
BenchmarkSpinLock_ReadLock/sync.Mutex-8                	88477630	        13.49 ns/op
BenchmarkSpinLock_ReadLock/sync.Mutex_2_threads
BenchmarkSpinLock_ReadLock/sync.Mutex_2_threads-8      	13087987	        94.88 ns/op
BenchmarkSpinLock_ReadLock/sync.Mutex_4_threads
BenchmarkSpinLock_ReadLock/sync.Mutex_4_threads-8      	 4595138	       259.6 ns/op
BenchmarkSpinLock_ReadLock/sync.Mutex_8_threads
BenchmarkSpinLock_ReadLock/sync.Mutex_8_threads-8      	 1906975	       630.7 ns/op
BenchmarkSpinLock_ReadLock/sync.RWMutex
BenchmarkSpinLock_ReadLock/sync.RWMutex-8              	84876637	        13.79 ns/op
BenchmarkSpinLock_ReadLock/sync.RWMutex_2_threads
BenchmarkSpinLock_ReadLock/sync.RWMutex_2_threads-8    	17766799	        69.35 ns/op
BenchmarkSpinLock_ReadLock/sync.RWMutex_4_threads
BenchmarkSpinLock_ReadLock/sync.RWMutex_4_threads-8    	 6728850	       177.8 ns/op
BenchmarkSpinLock_ReadLock/sync.RWMutex_8_threads
BenchmarkSpinLock_ReadLock/sync.RWMutex_8_threads-8    	 1740376	       688.3 ns/op
*/
func BenchmarkSpinLock_ReadLock(b *testing.B) {
	var (
		rw = true
	)
	b.Run("Spinlock 1 thread", func(b *testing.B) {
		l := new(Mutex)
		for i := 0; i < b.N; i++ {
			l.Lock()
			l.Unlock()
		}
	})

	b.Run("Spinlock 2 threads", func(b *testing.B) {
		var (
			l  = new(Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("Spinlock 4 threads", func(b *testing.B) {
		var (
			l  = new(Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("Spinlock 8 threads", func(b *testing.B) {
		var (
			l  = new(Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})

	if rw {
		b.Run("RWSpinlock 1 thread", func(b *testing.B) {
			l := new(RWMutex)
			for i := 0; i < b.N; i++ {
				l.RLock()
				l.RUnlock()
			}
		})

		b.Run("RWSpinlock 2 threads", func(b *testing.B) {
			var (
				l  = new(RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.RLock()
						l.RUnlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("RWSpinlock 4 threads", func(b *testing.B) {
			var (
				l  = new(RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.RLock()
						l.RUnlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("RWSpinlock 8 threads", func(b *testing.B) {
			var (
				l  = new(RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.RLock()
						l.RUnlock()
					}
				}()
			}
			wg.Wait()
		})
	}

	b.Run("sync.Mutex", func(b *testing.B) {
		l := new(sync.Mutex)
		for i := 0; i < b.N; i++ {
			l.Lock()
			l.Unlock()
		}
	})
	b.Run("sync.Mutex 2 threads", func(b *testing.B) {
		var (
			l  = new(sync.Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 2; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("sync.Mutex 4 threads", func(b *testing.B) {
		var (
			l  = new(sync.Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 4; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})
	b.Run("sync.Mutex 8 threads", func(b *testing.B) {
		var (
			l  = new(sync.Mutex)
			wg = new(sync.WaitGroup)
		)
		for i := 0; i < 8; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for i := 0; i < b.N; i++ {
					l.Lock()
					l.Unlock()
				}
			}()
		}
		wg.Wait()
	})

	if rw {
		b.Run("sync.RWMutex", func(b *testing.B) {
			var (
				l = new(sync.RWMutex)
			)
			for i := 0; i < b.N; i++ {
				l.RLock()
				l.RUnlock()
			}
		})
		b.Run("sync.RWMutex 2 threads", func(b *testing.B) {
			var (
				l  = new(sync.RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 2; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.RLock()
						l.RUnlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("sync.RWMutex 4 threads", func(b *testing.B) {
			var (
				l  = new(sync.RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 4; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.RLock()
						l.RUnlock()
					}
				}()
			}
			wg.Wait()
		})
		b.Run("sync.RWMutex 8 threads", func(b *testing.B) {
			var (
				l  = new(sync.RWMutex)
				wg = new(sync.WaitGroup)
			)
			for i := 0; i < 8; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					for i := 0; i < b.N; i++ {
						l.RLock()
						l.RUnlock()
					}
				}()
			}
			wg.Wait()
		})
	}
}

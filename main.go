package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/reactivex/rxgo/v2"
)

func t1() {
	observable := rxgo.Just(1, 2, 3, 4, 5)()
	ch := observable.Observe()
	for item := range ch {
		fmt.Println(item.V)
	}
}

func t2() {
	observable := rxgo.Just(1, 2, errors.New("unknown"), 4, 5)()
	ch := observable.Observe()
	for item := range ch {
		if item.Error() {
			fmt.Println("Error:", item.E)
		} else {
			fmt.Println(item.V)
		}
	}
}

func t3() {
	observable := rxgo.Just(1, 2, 3, 4, 5)()
	<-observable.ForEach(func(v interface{}) {
		fmt.Println("onNext:", v)
	}, func(err error) {
		fmt.Println("onError:", err)
	}, func() {
		fmt.Println("onComplete")
	})
}

func t4() {
	observable := rxgo.Create([]rxgo.Producer{
		func(ctx context.Context, next chan<- rxgo.Item) {
			next <- rxgo.Of(1)
		},
		func(ctx context.Context, next chan<- rxgo.Item) {
			next <- rxgo.Of(2)
			next <- rxgo.Error(errors.New("unknown"))
			next <- rxgo.Of(4)
			next <- rxgo.Of(5)
		},
	})
	<-observable.ForEach(func(v interface{}) {
		fmt.Println("onNext:", v)
	}, func(err error) {
		fmt.Println("onError:", err)
	}, func() {
		fmt.Println("onComplete")
	})
}

func t5() {
	ch := make(chan rxgo.Item) // no buffer
	go func() {
		for i := 1; i <= 5; i++ {
			ch <- rxgo.Of(i)
		}
		close(ch)
	}()

	observable := rxgo.FromChannel(ch)
	<-observable.ForEach(func(v interface{}) {
		fmt.Println("onNext:", v)
	}, func(err error) {
		fmt.Println("onError:", err)
	}, func() {
		fmt.Println("onComplete")
	})
}

func t6() {
	// Interval以传入的时间间隔生成一个无穷的数字序列，从 0 开始
	observable := rxgo.Interval(rxgo.WithDuration(3 * time.Second))
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t7() {
	// 生成一个范围内的数字，达到最大值就结束
	observable := rxgo.Range(0, 5)
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t8() {
	observable := rxgo.Just(1, 2, 3)()
	// 每隔指定时间，重复一次该序列，一共重复指定次数
	observable = observable.Repeat(5, rxgo.WithDuration(5*time.Second))
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t9() {
	observable := rxgo.Start([]rxgo.Supplier{
		func(ctx context.Context) rxgo.Item {
			return rxgo.Of(1)
		},
		func(ctx context.Context) rxgo.Item {
			return rxgo.Of(2)
		},
		func(ctx context.Context) rxgo.Item {
			return rxgo.Of(3)
		},
	})
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t10() {
	// Defer does not create the Observable until the observer subscribes,
	// and creates a fresh Observable for each observer.
	//
	// Cold Observable: 也就是在 subscribe 的时候才去生产数据流；
	// 与之相反的是 Hot Observable，也就是在创建 Observable 的时候就同时创建了数据流，
	// 这样第一次 subscribe 的时候就把数据消耗完了，再次 subscribe 是没有数据的，前面的
	// 例子都是 Hot Observable.
	observable := rxgo.Defer([]rxgo.Producer{func(_ context.Context, ch chan<- rxgo.Item) {
		for i := 0; i < 3; i++ {
			ch <- rxgo.Of(i)
		}
	}})

	// 有数据
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
	// 有数据
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t11() {
	ch := make(chan rxgo.Item)
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- rxgo.Of(i)
		}
		close(ch)
	}()

	// 普通的 Observable，只要有一个观察者注册成功就会释放数据
	observable := rxgo.FromChannel(ch)

	// 注册观察者1
	observable.DoOnNext(func(i interface{}) {
		fmt.Printf("First observer: %d\n", i)
	})

	time.Sleep(3 * time.Second)
	fmt.Println("before subscribe second observer")

	// 注册观察者2
	observable.DoOnNext(func(i interface{}) {
		fmt.Printf("Second observer: %d\n", i)
	})

	time.Sleep(3 * time.Second)
}

func t12() {
	ch := make(chan rxgo.Item)
	go func() {
		for i := 1; i <= 3; i++ {
			ch <- rxgo.Of(i)
		}
		close(ch)
	}()

	// 可连接的 Observable
	observable := rxgo.FromChannel(ch, rxgo.WithPublishStrategy())

	// 注册观察者1
	observable.DoOnNext(func(i interface{}) {
		fmt.Printf("First observer: %d\n", i)
	})

	time.Sleep(3 * time.Second)
	fmt.Println("before subscribe second observer")

	// 注册观察者2
	observable.DoOnNext(func(i interface{}) {
		fmt.Printf("Second observer: %d\n", i)
	})

	// 通知 Observable，意味着所有的观察者已经注册完毕，才开始释放数据
	// 另外，可连接的 Observable 是 Cold Observable，即每个观察者都会收到一份相同的拷贝。
	observable.Connect(context.Background())

	time.Sleep(5 * time.Second)

	fmt.Println("over.")
}

func t13() {
	observable := rxgo.Just(1, 2, 3)()
	// Map 转换或者过滤
	// 如果出现一个 error，整个数据流都是无效的
	observable = observable.Map(func(ctx context.Context, v interface{}) (interface{}, error) {
		// vv := v.(int)
		// if vv%2 == 0 {
		// 	return vv * 2, nil
		// } else {
		// 	return vv, errors.New("Error")
		// }

		return v.(int) * 2, nil
	}).Map(func(ctx context.Context, v interface{}) (interface{}, error) {
		return v.(int) + 1, nil
	})

	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func t14() {
	observable := rxgo.Just(
		User{
			Name: "dj",
			Age:  18,
		},
		User{
			Name: "jw",
			Age:  20,
		},
	)()

	observable = observable.Marshal(json.Marshal)

	for item := range observable.Observe() {
		fmt.Println(string(item.V.([]byte)))
	}
}

func t15() {
	observable := rxgo.Just(
		`{"name":"dj","age":18}`,
		`{"name":"jw","age":20}`,
	)()

	observable = observable.Map(func(_ context.Context, i interface{}) (interface{}, error) {
		return []byte(i.(string)), nil
	}).Unmarshal(json.Unmarshal, func() interface{} {
		return &User{}
	})

	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t16() {
	observable := rxgo.Just(1, 2, 3, 4)().BufferWithCount(3)
	for item := range observable.Observe() {
		fmt.Println(item.V) // item.V 此时是切片类型
	}
}

func t17() {
	count := 3

	observable := rxgo.Range(0, 10).GroupBy(count, func(item rxgo.Item) int {
		return item.V.(int) % count
	}, rxgo.WithBufferedChannel(10))

	for subObservable := range observable.Observe() {
		fmt.Println("New observable:")

		for item := range subObservable.V.(rxgo.Observable).Observe() {
			fmt.Printf("item: %v\n", item.V)
		}
	}
}

func t18() {
	observable := rxgo.Range(1, 20)

	observable = observable.Map(func(_ context.Context, i interface{}) (interface{}, error) {
		time.Sleep(time.Duration(rand.Int31()))
		return i.(int)*2 + 1, nil
	}, rxgo.WithCPUPool())

	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t19() {
	observable := rxgo.Range(1, 10)

	observable = observable.Filter(func(i interface{}) bool {
		return i.(int)%2 == 0
	})

	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t20() {
	observable := rxgo.Just(0, 1, 2, 3, 4)().ElementAt(2)

	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t21() {
	ch := make(chan rxgo.Item)

	go func() {
		ch <- rxgo.Of(1)
		time.Sleep(2 * time.Second)
		ch <- rxgo.Of(2)
		ch <- rxgo.Of(3)
		time.Sleep(2 * time.Second)
		close(ch)
	}()

	observable := rxgo.FromChannel(ch).Debounce(rxgo.WithDuration(1 * time.Second))
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t22() {
	observable := rxgo.Just(1, 2, 2, 3, 3, 4, 4)().
		Distinct(func(_ context.Context, i interface{}) (interface{}, error) {
			return i, nil
		})
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t23() {
	// 跳过前若干个数据
	observable := rxgo.Just(1, 2, 3, 4, 5)().Skip(2)
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func t24() {
	// 只取前若干个数据
	observable := rxgo.Just(1, 2, 3, 4, 5)().Take(2)
	for item := range observable.Observe() {
		fmt.Println(item.V)
	}
}

func main() {
	t17()
}

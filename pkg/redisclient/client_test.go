package redisclient

import (
	"fmt"
	"strconv"
	"testing"

	"math/rand"

	"github.com/linkai-io/am/pkg/convert"

	"github.com/gomodule/redigo/redis"
)

func BenchmarkGetKeys(b *testing.B) {
	r := New("0.0.0.0:6379", "test132")
	if err := r.Init(); err != nil {
		b.Fatalf("unable to connect: %s\n", err)
	}
	key := "test:benchget"
	conn := r.Get()
	defer conn.Close()

	testInsertSetKeys(conn, key, 100000, b)
	sample1 := testGetRandomSample(conn, key, 1000, b)
	sample2 := testGetRandomSample(conn, key, 10000, b)

	found := 0
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		data, err := redis.Strings(redis.Values(conn.Do("SMEMBERS", key)))
		if err != nil {
			b.Fatalf("error getting members: %s\n", err)
		}
		members := make(map[string]struct{}, len(data))
		for j := 0; j < len(data); j++ {
			members[data[j]] = struct{}{}
		}

		for j := 0; j < len(sample1); j++ {
			if _, ok := members[sample1[j]]; ok {
				found++
			}
		}

		for j := 0; j < len(sample2); j++ {
			if _, ok := members[sample2[j]]; ok {
				found++
			}
		}
	}
	b.Logf("Found: %d\n", found)
}

func BenchmarkIntersect(b *testing.B) {
	r := New("0.0.0.0:6379", "test132")
	if err := r.Init(); err != nil {
		b.Fatalf("unable to connect: %s\n", err)
	}
	key := "test:benchintersec"
	conn := r.Get()
	defer conn.Close()

	testInsertSetKeys(conn, key, 100000, b)
	sample1 := testGetRandomSample(conn, key, 1000, b)
	sample2 := testGetRandomSample(conn, key, 10000, b)
	found := 0

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		sampleKey := strconv.FormatInt(rand.Int63(), 10)
		args := make([]interface{}, len(sample1)+1)
		args[0] = sampleKey
		for i := 0; i < len(sample1); i++ {
			args[i+1] = sample1[i]
		}
		if _, err := conn.Do("SADD", args...); err != nil {
			b.Fatalf("error adding")
		}
		data, err := redis.Strings(redis.Values(conn.Do("SINTER", key, sampleKey)))
		if err != nil {
			b.Fatalf("error getting random sample: %s\n", err)
		}

		found += len(data)
		if _, err := conn.Do("DEL", sampleKey); err != nil {
			b.Fatalf("error deleting key: %s\n", err)
		}

		sampleKey = strconv.FormatInt(rand.Int63(), 10)
		args = make([]interface{}, len(sample2)+1)
		args[0] = sampleKey
		for i := 0; i < len(sample2); i++ {
			args[i+1] = sample2[i]
		}
		if _, err := conn.Do("SADD", args...); err != nil {
			b.Fatalf("error adding")
		}

		data, err = redis.Strings(redis.Values(conn.Do("SINTER", key, sampleKey)))
		if err != nil {
			b.Fatalf("error getting random sample: %s\n", err)
		}
		if _, err := conn.Do("DEL", sampleKey); err != nil {
			b.Fatalf("error deleting key: %s\n", err)
		}
		found += len(data)
	}
	b.Logf("%d found\n", found)
}

func testGetRandomSample(conn redis.Conn, key string, count int, b *testing.B) []string {
	data, err := redis.Strings(redis.Values(conn.Do("SRANDMEMBER", key, count)))
	if err != nil {
		b.Fatalf("error getting random sample: %s\n", err)
	}
	return data
}

func testInsertSetKeys(conn redis.Conn, key string, count int, b *testing.B) {
	if err := conn.Send("MULTI"); err != nil {
		b.Fatalf("error sending multi command: %s\n", err)
	}

	for i := 0; i < count; i++ {
		addr := fmt.Sprintf("192.168.0.%d", i)
		host := fmt.Sprintf("%d.somedomain.com", i)
		hash := convert.HashAddress(addr, host)
		if err := conn.Send("SADD", key, hash); err != nil {
			b.Fatalf("error sending sadd: %s\n", err)
		}
	}

	if _, err := conn.Do("EXEC"); err != nil {
		b.Fatalf("error sending multi command: %s\n", err)
	}
}

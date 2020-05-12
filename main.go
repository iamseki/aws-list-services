package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/rds"
	_ "github.com/joho/godotenv/autoload"
)

// Instance for json structure
type Instance struct {
	Name     string
	Endpoint map[string]interface{}
	Engine   string
}

// updateCredentials to receive specific aws service data
func updateCredentials(sess *session.Session) *aws.Config {
	var region = os.Getenv("region")
	var role = os.Getenv("role")

	assumeRole := os.Getenv("role_" + role)

	creds := stscreds.NewCredentials(sess, assumeRole)

	return &aws.Config{Credentials: creds, Region: &region}
}

// writeToLogFile
func writeToLogFile(text string) {
	f, err := os.OpenFile("instances.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
	}
	defer f.Close()
	if _, err := f.WriteString(text); err != nil {
		log.Println(err)
	}
}

// jsonStrigify to formatted and indent json data and return in string format
func jsonStringify(instance Instance) string {
	b, _ := json.MarshalIndent(instance, "", "    ")

	return string(b)
}

// listRds instances
func listRds(sess *session.Session, c chan []*rds.DBInstance, wg *sync.WaitGroup) {
	defer wg.Done()
	svc := rds.New(sess, updateCredentials(sess))

	if result, err := svc.DescribeDBInstances(nil); err == nil {
		c <- result.DBInstances
	}
}

// handleRds call jsonSringify and prints it result
func handleRds(c chan []*rds.DBInstance, wg *sync.WaitGroup) {
	defer wg.Done()
	var rdsInstances []string

	for _, d := range <-c {
		instance := Instance{*d.DBInstanceIdentifier,
			map[string]interface{}{"Address": *d.Endpoint.Address, "Port": *d.Endpoint.Port},
			*d.Engine + ":" + *d.EngineVersion,
		}

		rdsInstances = append(rdsInstances, jsonStringify(instance))
	}

	close(c)
	rdsLen := len(rdsInstances)
	capac := fmt.Sprintf("[%d] RDS Instances: \n", rdsLen)
	writeToLogFile(capac)
	fmt.Printf("[%d] RDS Instances : \n", rdsLen)

	for _, instance := range rdsInstances {
		fmt.Println(instance + ",")
		writeToLogFile(instance + ",\n")
	}

}

// listElastic instances
func listElastic(sess *session.Session, c chan []*elasticache.CacheCluster, wg *sync.WaitGroup) {
	defer wg.Done()
	svc := elasticache.New(sess, updateCredentials(sess))

	sni := true

	if result, err := svc.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
		ShowCacheNodeInfo:                       &sni,
		ShowCacheClustersNotInReplicationGroups: &sni}); err == nil {
		c <- result.CacheClusters
	}

}

// listElasticReplicas instances
func listElasticReplicas(sess *session.Session, c chan []*elasticache.ReplicationGroup, wg *sync.WaitGroup) {
	defer wg.Done()
	svc := elasticache.New(sess, updateCredentials(sess))

	if result, err := svc.DescribeReplicationGroups(nil); err == nil {
		c <- result.ReplicationGroups
	}
}

// init env vars & file log
func init() {
	go os.Remove("instances.log")
}

// main is the entry point to our program
func main() {
	var wg sync.WaitGroup

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	}))

	result, elastic, replicas :=
		make(chan []*rds.DBInstance),
		make(chan []*elasticache.CacheCluster),
		make(chan []*elasticache.ReplicationGroup)

	wg.Add(6)

	go listRds(sess, result, &wg)

	go listElastic(sess, elastic, &wg)

	go listElasticReplicas(sess, replicas, &wg)

	go handleRds(result, &wg)

	var elastiInstances []string

	go func(c chan []*elasticache.CacheCluster, eSlice *[]string, wg *sync.WaitGroup) {
		defer wg.Done()
		for _, e := range <-c {
			instance := Instance{*e.CacheClusterId,
				map[string]interface{}{"Address": *e.CacheNodes[0].Endpoint.Address, "Port": *e.CacheNodes[0].Endpoint.Port},
				*e.Engine + ":" + *e.EngineVersion,
			}

			*eSlice = append(*eSlice, jsonStringify(instance))
		}
		close(c)
	}(elastic, &elastiInstances, &wg)

	go func(c chan []*elasticache.ReplicationGroup, rSlice *[]string, wg *sync.WaitGroup) {
		defer wg.Done()
		for _, e := range <-c {
			instance := Instance{*e.ReplicationGroupId,
				map[string]interface{}{"Address": *e.NodeGroups[0].PrimaryEndpoint.Address, "Port": *e.NodeGroups[0].PrimaryEndpoint.Port},
				"redis",
			}

			*rSlice = append(*rSlice, jsonStringify(instance))
		}
		close(c)
	}(replicas, &elastiInstances, &wg)

	wg.Wait()

	eLen := len(elastiInstances)

	capac := fmt.Sprintf("[%d] ElastiCache Instances: \n", eLen)
	writeToLogFile(capac)

	fmt.Printf("[%d] Elasti Instances:\n", len(elastiInstances))
	for _, instance := range elastiInstances {
		fmt.Println(instance + ",")
		writeToLogFile(instance + ",\n")
	}
}

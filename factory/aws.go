package factory

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/aws/aws-sdk-go/service/rds"

	// load env vars
	_ "github.com/joho/godotenv/autoload"
)

var sess *session.Session

// AWS blabla
type AWS interface {
	List(*sync.WaitGroup)
}

type instance struct {
	Name     string
	Endpoint map[string]interface{}
	Engine   string
}

type rdsService struct {
	svc *rds.RDS
}

type elastiService struct {
	svc *elasticache.ElastiCache
}

func (r rdsService) List(wg *sync.WaitGroup) {
	defer wg.Done()

	if result, err := r.svc.DescribeDBInstances(nil); err == nil {
		var rdsi []string

		for _, db := range result.DBInstances {
			i := instance{*db.DBInstanceIdentifier,
				map[string]interface{}{"Address": *db.Endpoint.Address, "Port": *db.Endpoint.Port},
				*db.Engine + ":" + *db.EngineVersion,
			}
			rdsi = append(rdsi, jsonStringify(i))
		}

		capac := fmt.Sprintf("[%d] RDS Instances: ", len(rdsi))

		rdsi = append([]string{capac}, rdsi...)

		for _, i := range rdsi {
			fmt.Println(i)
		}

		wg.Add(1)
		go writeLog("rdsInstances.log", rdsi, wg)

		return
	}
}

func (e elastiService) List(wg *sync.WaitGroup) {
	defer wg.Done()

	elastic, replicas :=
		make(chan []*elasticache.CacheCluster),
		make(chan []*elasticache.ReplicationGroup)

	go listElasti(*e.svc, elastic)

	go listElastiRepl(*e.svc, replicas)

	esi := elastiWorker(elastic, replicas)

	capac := fmt.Sprintf("[%d] Elasti Instances: ", len(esi))

	esi = append([]string{capac}, esi...)

	wg.Add(1)
	go writeLog("elastinstances.log", esi, wg)

	for _, i := range esi {
		fmt.Println(i)
	}
}

func elastiWorker(e chan []*elasticache.CacheCluster, r chan []*elasticache.ReplicationGroup) []string {
	var elastiInstances []string

	for i := 0; i < 2; i++ {
		select {
		case dbi := <-e:
			for _, elast := range dbi {
				d := instance{*elast.CacheClusterId,
					map[string]interface{}{"Address": *elast.CacheNodes[0].Endpoint.Address,
						"Port": *elast.CacheNodes[0].Endpoint.Port},
					*elast.Engine + ":" + *elast.EngineVersion,
				}

				elastiInstances = append(elastiInstances, jsonStringify(d))
			}
		case dbr := <-r:
			for _, repl := range dbr {
				d := instance{*repl.ReplicationGroupId,
					map[string]interface{}{"Address": *repl.NodeGroups[0].PrimaryEndpoint.Address,
						"Port": *repl.NodeGroups[0].PrimaryEndpoint.Port},
					"redis",
				}

				elastiInstances = append(elastiInstances, jsonStringify(d))
			}
		}
	}

	return elastiInstances
}

// listElastic instances
func listElasti(svc elasticache.ElastiCache, c chan []*elasticache.CacheCluster) {

	sni := true

	if result, err := svc.DescribeCacheClusters(&elasticache.DescribeCacheClustersInput{
		ShowCacheNodeInfo:                       &sni,
		ShowCacheClustersNotInReplicationGroups: &sni}); err == nil {
		c <- result.CacheClusters
	}
}

// listElasticReplicas instances
func listElastiRepl(svc elasticache.ElastiCache, c chan []*elasticache.ReplicationGroup) {
	if result, err := svc.DescribeReplicationGroups(nil); err == nil {
		c <- result.ReplicationGroups
	}
}

func updateCredentials(sess *session.Session) *aws.Config {
	var region = os.Getenv("region")
	var role = os.Getenv("role")

	assumeRole := os.Getenv("role_" + role)

	creds := stscreds.NewCredentials(sess, assumeRole)

	return &aws.Config{Credentials: creds, Region: &region}
}

func jsonStringify(i instance) string {
	b, _ := json.MarshalIndent(i, "", "    ")

	return string(b)
}

func writeLog(file string, text []string, wg *sync.WaitGroup) {
	defer wg.Done()

	f, err := os.OpenFile(file, os.O_TRUNC|os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	if err != nil {
		log.Println(err)
	}

	for _, t := range text {
		f.WriteString(t)
	}
}

func init() {
	sess = session.Must(session.NewSession())
}

// AWSFactory Generates AWS services
func AWSFactory(service string) AWS {
	service = strings.ToLower(service)
	switch service {
	case "rds":
		return rdsService{
			svc: rds.New(sess, updateCredentials(sess)),
		}
	case "elasti":
		return elastiService{
			svc: elasticache.New(sess, updateCredentials(sess)),
		}
	default:
		return nil
	}
}

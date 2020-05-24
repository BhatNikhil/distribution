package encode

import (
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/golang-collections/collections/set"
	"github.com/opencontainers/go-digest"
)

//InsertNodeAsSet removes the set if already exists and inserts the new set
func (emngr *EncodeManager) InsertNodeAsSet(nodeID string, keys []string) {
	conn := emngr.redisPool.Get()
	defer conn.Close()

	nodeSetKey := getNodeSetKey(nodeID)
	conn.Do("DEL", nodeSetKey)
	emngr.BulkInsertSet(nodeSetKey, keys)

}

//GetAvailableBlocksFromNode will get the instersection of the node and the recipe
func (emngr *EncodeManager) GetAvailableBlocksFromNode(nodeID string, digest digest.Digest) (set.Set, error) {
	conn := emngr.redisPool.Get()
	defer conn.Close()

	nodeSetKey := getNodeSetKey(nodeID)
	if exists, _ := redis.Bool(conn.Do("EXISTS", nodeSetKey)); !exists {
		fmt.Println("Node info not available in server stash")
		return set.Set{}, fmt.Errorf("Node info not available in server stash")
	}

	recipeSetKey := getRecipeSetKey(digest)

	sum := recipeSetKey + "+" + nodeSetKey
	conn.Do("SINTERSTORE", sum, nodeSetKey, recipeSetKey)
	values, _ := redis.Strings(conn.Do("SMEMBERS", sum))
	if values == nil || len(values) == 0 {
		fmt.Printf("Result of SINTER for digest %s is EMPTY\n", digest)
		return set.Set{}, nil
	}

	setValues := make([]interface{}, len(values))
	for i, v := range values {
		setValues[i] = v
	}

	fmt.Printf("serverless==> For digest: %s Result of SINTER: %d \n", digest, len(setValues))
	return *set.New(setValues...), nil
}

func getNodeSetKey(nodeID string) string {
	return "node-set:" + nodeID
}

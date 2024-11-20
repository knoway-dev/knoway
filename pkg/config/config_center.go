package config

import (
	"encoding/json"
	"log"
	"os"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	v1alpha4 "knoway.dev/api/clusters/v1alpha1"
)

type ConfigsServer struct {
	Clusters map[string]*v1alpha4.Cluster `json:"clusters"`
	mu       sync.RWMutex
	filePath string
	changed  bool // changed flag to indicate if a save is needed
}

func NewConfigsServer(filePath string) *ConfigsServer {
	return &ConfigsServer{
		Clusters: make(map[string]*v1alpha4.Cluster),
		filePath: filePath,
		changed:  false,
	}
}

func (cs *ConfigsServer) LoadFromFile() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	// Check if the file exists, if not create it
	if _, err := os.Stat(cs.filePath); os.IsNotExist(err) {
		if err := os.WriteFile(cs.filePath, []byte("{}"), 0644); err != nil {
			log.Printf("Failed to create config file: %v", err)
			return err
		}
		log.Printf("Config file created: %s", cs.filePath)
	}

	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		log.Printf("Failed to read config file: %v", err)
		return err
	}
	log.Printf("Loaded config from file: %s", cs.filePath)
	return json.Unmarshal(data, &cs.Clusters)
}

func (cs *ConfigsServer) SaveToFile() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	data, err := json.Marshal(&cs.Clusters)
	if err != nil {
		log.Printf("Failed to marshal config data: %v", err)
		return err
	}
	if err := os.WriteFile(cs.filePath, data, 0644); err != nil {
		log.Printf("Failed to save config to file: %v", err)
		return err
	}
	log.Printf("Saved config to file: %s", cs.filePath)
	return nil
}

func (cs *ConfigsServer) Start() {
	log.Printf("Starting ConfigServer at %s", cs.filePath)
	// Clear the file at startup to avoid old history
	if err := os.WriteFile(cs.filePath, []byte("{}"), 0644); err != nil {
		log.Printf("Failed to initialize config file: %v", err)
	} else {
		log.Printf("Initialized config file: %s", cs.filePath)
	}

	if err := cs.LoadFromFile(); err != nil {
		log.Printf("Error loading config from file: %v", err)
	}
	go cs.scheduleSave() // Start the scheduled save routine
}

func (cs *ConfigsServer) scheduleSave() {
	for {
		time.Sleep(3 * time.Second) // Wait for 3 seconds
		cs.mu.Lock()
		changed := cs.changed
		cs.mu.Unlock()

		if changed {
			if err := cs.SaveToFile(); err != nil {
				log.Printf("Error saving config: %v", err)
			}

			cs.mu.Lock()
			cs.changed = false // Reset the changed flag after saving
			cs.mu.Unlock()
		}
	}
}

func (cs *ConfigsServer) UpsertCluster(name string, cluster *v1alpha4.Cluster) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.Clusters[name] = cluster
	cs.changed = true // Mark as changed
	log.Printf("Upserted cluster: %s", name)
}

func (cs *ConfigsServer) RemoveCluster(name string) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.Clusters, name)
	cs.changed = true // Mark as changed
	log.Printf("Removed cluster: %s", name)
}

func (cs *ConfigsServer) GetCluster(name string) (*v1alpha4.Cluster, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	cluster, exists := cs.Clusters[name]
	log.Printf("Retrieved cluster: %s, exists: %v", name, exists)
	return cluster, exists
}

func (cs *ConfigsServer) GetClusters() map[string]*v1alpha4.Cluster {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	log.Printf("Retrieved all clusters, total count: %d", len(cs.Clusters))
	return cs.Clusters
}

type ConfigClient struct {
	server          *ConfigsServer
	changedCallFunc func(clusters map[string]*v1alpha4.Cluster) // User-defined function to handle changes
}

func NewConfigClient(filePath string, changedCallFunc func(clusters map[string]*v1alpha4.Cluster)) *ConfigClient {
	server := &ConfigsServer{filePath: filePath} // Initialize ConfigsServer
	return &ConfigClient{
		server:          server,
		changedCallFunc: changedCallFunc,
	}
}

func (cc *ConfigClient) Start() {
	cc.LoadFromFile()
	go cc.watchFile() // Start watching the file for changes
}

func (cc *ConfigClient) watchFile() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(cc.server.filePath)
	if err != nil {
		log.Fatalf("Failed to add file to watcher: %v", err)
	}

	for {
		select {
		case _, ok := <-watcher.Events:
			if !ok {
				return
			}
			log.Printf("File changed, reloading clusters from file: %s", cc.server.filePath)
			cc.LoadFromFile()
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Watcher error: %v", err)
		}
	}
}

func (cc *ConfigClient) LoadFromFile() {
	if err := cc.server.LoadFromFile(); err != nil {
		log.Printf("Error reloading clusters from file: %v", err)
	}
	if cc.changedCallFunc != nil {
		// Call the user-defined function if it exists
		cc.changedCallFunc(cc.GetClusters()) // Pass appropriate action and cluster
	}
}

func (cc *ConfigClient) GetCluster(name string) (*v1alpha4.Cluster, bool) {
	return cc.server.GetCluster(name)
}

func (cc *ConfigClient) GetClusters() map[string]*v1alpha4.Cluster {
	return cc.server.GetClusters()
}

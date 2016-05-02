package beego_cache

import (
  "github.com/astaxie/beego/cache"
  _ "github.com/astaxie/beego/cache/redis"
  "github.com/astaxie/beego"
  // "github.com/astaxie/beego/logs"
  "os"
  "time"
  "fmt"
  "strconv"
  "strings"
)

type Cache struct {
  cache     cache.Cache
}

var instantiated *Cache = nil
const PREFIX = "[ CACHE ] "

func GetCache() *Cache {
  if instantiated == nil {
    instantiated = new(Cache);
    instantiated.CacheInit()
  }
  return instantiated;
}

func (this *Cache) FileChanged(path string) bool {
  cache_size, cache_time, cache_new_file, err := this.FileCacheStat(path)
  // Debug("cache %d, %s", cache_size, cache_time)
  if err != nil {
    Error("Filecachestat error: %s", err.Error())
    return true
  }
  
  fi, err := os.Stat(path)
  if err != nil {
    Warning("Unable to stat asset %s", path)// Could not obtain stat, handle error
    return true
  }
  
  if fi.Size() != cache_size || fi.ModTime() != cache_time  {
    return true
  }
  
  //ensure file exists
  if (cache_new_file != "") {
    if _, err := os.Stat(cache_new_file); os.IsNotExist(err) {
      Warning("Cache file missing %s", cache_new_file)
      return true
    }
  }
  
  return false
}

func (this *Cache) FileCacheStat(path string) (file_size int64, file_modtime time.Time, file_dest string, err error) {
  name := fmt.Sprintf("file_%s", path)
  if this.CacheExists(name) {
    cache := this.CacheGet(name)
    res   := strings.Split(cache, "|")
    file_size, err := strconv.ParseInt(res[0], 10, 64)
    if (err != nil) {
      Warning("FileCacheStat: can't get size from cache for %s", path)
    }
    // https://gobyexample.com/time-formatting-parsing
    file_modtime, err := time.Parse(time.RFC3339, res[1]) //TODO
    if (err != nil) {
      Warning("FileCacheStat: mod time from cache: %s: %s", res[1], err.Error())
    }
    file_dest  := res[2]
    return file_size, file_modtime, file_dest, nil
  } else {
    Debug("Cache missing %s", name)
    return 0, time.Time{}, "", fmt.Errorf("File missing: %s", path)
  }
}

func (this *Cache) CacheFile(path string, stat os.FileInfo, new_file_path string) error {
  name := fmt.Sprintf("file_%s", path)
  var size int64 = 0
  if (stat == nil) {
    stat2, err := os.Stat(path)
    if (err != nil) {
      Warning("Unable to stat file %s, %s", path, err.Error())
      return err
    }
    stat = stat2
  }
  size = stat.Size()
  err := this.CacheSetMax(name, fmt.Sprintf("%d|%s|%s", size, stat.ModTime().Format(time.RFC3339), new_file_path))
  if err != nil {
    Error("CacheFile error %s", err.Error())
    return err
  }
  Debug("cached %s to %s", path, new_file_path)
  return nil
}

func (this *Cache) CacheInit() error {
  if this.cache == nil {
    cache_engine := beego.AppConfig.DefaultString("cache_engine", "file")
    cache_config := beego.AppConfig.DefaultString("cache_config", "")
    if (cache_engine == "file" && cache_config == "") {
      cache_config = `{"CachePath":"./tmp/cache","FileSuffix":".cache","DirectoryLevel":2,"EmbedExpiry":86400}`
    }
    if (cache_engine == "redis" && cache_config == "") {
      cache_config = `{"conn":":6379"}`
    }
    if (cache_config == "memcache" && cache_config == "") {
      cache_config = `{"conn":"127.0.0.1:11211"}`
    }
    // Debug("cache: %s %s", cache_engine, cache_config)
    bm, err := cache.NewCache(cache_engine, cache_config)  
    this.cache = bm
    if err != nil{
      Error("cache init error: %s", err.Error())
      return fmt.Errorf("cache init error: %s", err.Error())
    }
  }
  return nil
}

func (this *Cache) CacheName(name string) string {
  res := strings.Replace(name, "/", ":", -1)
  return res
}

func (this *Cache) CacheGet(name string) string {
  result := fmt.Sprintf("%s", this.cache.Get(this.CacheName(name)))
  //Debug("cache_get %s %q", name, result)
  return result
}

func (this *Cache) CacheSet(name string, value string, timeout time.Duration) error {
  err := this.cache.Put(this.CacheName(name), value, timeout)
  return err
}

func (this *Cache) CacheSetMax(name string, value string) error {
  return this.CacheSet(name, value, 60 * 60 * 24 * 365 * time.Second)
}

func (this *Cache) CacheExists(name string) bool {
  return this.cache.IsExist(this.CacheName(name))
}
func Debug(format string, v... interface{}) {
  beego.Debug(fmt.Sprintf(PREFIX + format, v...))
}
func Warning(format string, v... interface{}) {
  beego.Warning(fmt.Sprintf(PREFIX + format, v...))
}
func Error(format string, v... interface{}) {
  beego.Error(fmt.Sprintf(PREFIX + format, v...))
}

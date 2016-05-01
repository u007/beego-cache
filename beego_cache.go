package beego_cache

import (
  "github.com/astaxie/beego/cache"
	_ "github.com/astaxie/beego/cache/redis"
  "github.com/astaxie/beego"
  "github.com/astaxie/beego/logs"
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
var Logger = logs.NewLogger(10000)
const PREFIX = "[ CACHE ] "

func GetCache() *Cache {
  if instantiated == nil {
    instantiated = new(Cache);
    instantiated.cacheInit()
  }
  return instantiated;
}

func (this *Cache) fileChanged(path string) bool {
	cache_size, cache_time, cache_new_file, err := this.fileCacheStat(path)
	if err != nil {
		beego.Debug("filecachestat error:", err.Error())
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
	if _, err := os.Stat(cache_new_file); os.IsNotExist(err) {
		Warning("Cache file missing", cache_new_file)
		return true
	}
	
	return false
}

func (this *Cache) fileCacheStat(path string) (file_size int64, file_modtime time.Time, file_dest string, err error) {
	name := fmt.Sprintf("file_%s", path)
	if this.cache_exists(name) {
		// beego.Debug("Cache exists", name)
		cache := this.cache_get(name)
		res   := strings.Split(cache, "|")
		// beego.Debug(fmt.Sprintf("result: %q", res))
		file_size, err := strconv.ParseInt(res[0], 10, 64)
		if (err != nil) {
			Warning("fileCacheStat: can't get size from cache for %s", path)
		}
		// beego.Debug("cache:", cache)
		// https://gobyexample.com/time-formatting-parsing
		file_modtime, err := time.Parse(time.RFC3339, res[1]) //TODO
		if (err != nil) {
			Warning("fileCacheStat: mod time from cache: %s: %s", res[1], err.Error())
		}
		file_dest  := res[2]
		return file_size, file_modtime, file_dest, nil
	} else {
		beego.Debug("Cache missing", name)
		return 0, time.Time{}, "", fmt.Errorf("File missing: %f", path)
	}
}

func (this *Cache) cacheFile(path string, stat os.FileInfo, new_file_path string) {
	name := fmt.Sprintf("file_%s", path)
	err := this.cache_set_max(name, fmt.Sprintf("%d|%s|%s", stat.Size(), stat.ModTime().Format(time.RFC3339), new_file_path))
	if err != nil {
		beego.Debug("error set cacheFile", err.Error())
	}
}

func (this *Cache) cacheInit() error {
	if this.cache == nil {
		bm, err := cache.NewCache("redis", `{"conn":":6379"}`)  
		this.cache = bm
		if err != nil{
			beego.Error("cache init error: ", err.Error())
			return fmt.Errorf("cache init error: ", err.Error())
		}
	}
	return nil
}

func (this *Cache) cache_name(name string) string {
	res := strings.Replace(name, "/", ":", -1)
	return res
}

func (this *Cache) cache_get(name string) string {
	result := fmt.Sprintf("%s", this.cache.Get(this.cache_name(name)))
	// beego.Debug("cache_get", name, fmt.Sprintf("%q", result))
	return result
}

func (this *Cache) cache_set(name string, value string, timeout time.Duration) error {
	err := this.cache.Put(this.cache_name(name), value, timeout)
	return err
}

func (this *Cache) cache_set_max(name string, value string) error {
	return this.cache_set(name, value, 60 * 60 * 24 * 365 * time.Second)
}

func (this *Cache) cache_exists(name string) bool {
	return this.cache.IsExist(this.cache_name(name))
}

func Warning(format string, v... interface{}) {
	Logger.Warning(PREFIX + format, v...)
}
func Error(format string, v... interface{}) {
	Logger.Error(PREFIX + format, v...)
}

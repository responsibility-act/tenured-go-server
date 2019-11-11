package executors

import (
	"fmt"
	"github.com/ihaiker/tenured-go-server/commons"
	"errors"
	"regexp"
	"strconv"
)

type ExecutorManager interface {
	commons.Service

	Get(name string) ExecutorService
	Fix(name string, size, buffer int) ExecutorService
	Single(name string, buffer int) ExecutorService

	Config(config map[string]string) error
}

type defExecutorManager struct {
	def         ExecutorService
	executorMap map[string]ExecutorService
}

func (this *defExecutorManager) Get(module string) ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		return this.def
	}
}

func (this *defExecutorManager) Fix(module string, size, buffer int) ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		executor = NewFixedExecutorService(size, buffer)
		this.executorMap[module] = executor
		return executor
	}
}

func (this *defExecutorManager) Single(module string, buffer int) ExecutorService {
	if executor, has := this.executorMap[module]; has {
		return executor
	} else {
		executor = NewSingleExecutorService(buffer)
		this.executorMap[module] = executor
		return executor
	}
}

func executorParam(value string) (exeType string, param []int, err error) {
	m := regexp.MustCompile(`(fix|single|scheduled)\((\d+),?(\d+)?\)`)
	if m.MatchString(value) {
		gs := m.FindStringSubmatch(value)

		param := make([]int, len(gs[2:]))
		for i := 0; i < len(gs[2:]); i++ {
			param[i], _ = strconv.Atoi(gs[2+i])
		}
		return gs[1], param, nil
	} else {
		return "", nil, errors.New("执行定义错误: " + value)
	}
}

func (this *defExecutorManager) Config(config map[string]string) error {
	for executorName, configValue := range config {
		if _, has := this.executorMap[executorName]; !has {
			if execType, param, err := executorParam(configValue); err != nil {
				return err
			} else {
				switch execType {
				case "fix":
					this.Fix(execType, param[0], param[1])
				case "single":
					this.Single(execType, param[0])
				case "scheduled":
					//TODO 需要实现 scheduled queue
				default:
					return errors.New(fmt.Sprintf("未发现执行线程池定义方案：%s = %s", executorName, configValue))
				}
			}
		} else {
			return errors.New(fmt.Sprintf("执行线程池重复定义：%s", executorName))
		}
	}
	return nil
}

func (this *defExecutorManager) Start() error {
	return nil
}

func (this *defExecutorManager) Shutdown(interrupt bool) {
	for _, v := range this.executorMap {
		v.Shutdown(interrupt)
	}
}

func NewExecutorManager(def ExecutorService) ExecutorManager {
	return &defExecutorManager{
		def:         def,
		executorMap: map[string]ExecutorService{},
	}
}

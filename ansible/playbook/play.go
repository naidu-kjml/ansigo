package playbook

import (
  "reflect"
)

var play_fields = map[string]FieldAttribute{
  // fields used for validation only
  "pre_tasks": FieldAttribute{SkipLoad: true},
  "tasks": FieldAttribute{SkipLoad: true},
  "post_tasks": FieldAttribute{SkipLoad: true},
  "handlers": FieldAttribute{SkipLoad: true},
  "roles": FieldAttribute{SkipLoad: true},
  // fields pulled from YAML directly
  "hosts": FieldAttribute{T: "list", Default: nil, ListOf: "string"},
  "fact_path": FieldAttribute{T: "string", Default: nil},
  "gather_facts": FieldAttribute{T: "bool", Default: nil},
  "gather_subset": FieldAttribute{T:"barelist", Default: nil},
  "gather_timeout": FieldAttribute{T:"int", Default: nil},
  "vars_files": FieldAttribute{T:"list", Default: nil, Priority: 99},
  "vars_prompt": FieldAttribute{T:"list", Default: nil},
  "vault_password": FieldAttribute{T:"string", Default: nil},
  "force_handlers": FieldAttribute{T:"bool", Default: false},
  "max_fail_percentage": FieldAttribute{T:"float64", Default: 0.0},
  "serial": FieldAttribute{T:"list", Default: nil, ListOf: "int"},
  "strategy": FieldAttribute{T:"string", Default: "linear"},
  "order": FieldAttribute{T:"string", Default: nil},
}

type Play struct {
  Base
  Become
  Taggable

  // Non-yaml Attributes
  RemovedHosts map[string]bool
  // role attributes
  //roles []Role
  // block and task lists are read from yaml, but not via
  // the normal LoadValidFields method.
  Handlers []Block
  Pre_tasks []Block
  Tasks []Block
  Post_tasks []Block

  // Field attributes read from yaml
  Attr_hosts interface{}
  // facts
  Attr_fact_path interface{}
  Attr_gather_facts interface{}
  Attr_gather_subset interface{}
  Attr_gather_timeout interface{}
  // variable attributes
  Attr_vars_files interface{}
  Attr_vars_prompt interface{}
  Attr_vault_password interface{}
  // flag/setting attributes
  Attr_force_handlers interface{}
  Attr_max_fail_percentage interface{}
  Attr_serial interface{}
  Attr_strategy interface{}
  Attr_order interface{}
}

func (p *Play) GetAllObjectFieldAttributes() map[string]FieldAttribute {
  var all_fields = make(map[string]FieldAttribute)
  var items = []map[string]FieldAttribute{base_fields, taggable_fields, become_fields, play_fields}
  for i := 0; i < len(items); i++ {
    for k, v := range items[i] {
      all_fields[k] = v
    }
  }
  return all_fields
}

func (p *Play) GetInheritedValue(attr string) interface{} {
  field_name := "Attr_" + attr
  s := reflect.ValueOf(p).Elem()
  field := s.FieldByName(field_name)

  var cur_value interface{}
  if field.Kind() != reflect.Invalid {
    cur_value = field.Interface()
  } else {
    cur_value = nil
  }
  return cur_value
}

func (p *Play) Load(data map[interface{}]interface{}) {
  p.Base.Load(data)
  p.Taggable.Load(data)
  p.Become.Load(data)

  p.Base.GetInheritedValue = p.GetInheritedValue
  p.Base.GetAllObjectFieldAttributes = p.GetAllObjectFieldAttributes
  p.Taggable.GetInheritedValue = p.GetInheritedValue
  p.Taggable.GetAllObjectFieldAttributes = p.GetAllObjectFieldAttributes
  p.Become.GetInheritedValue = p.GetInheritedValue
  p.Become.GetAllObjectFieldAttributes = p.GetAllObjectFieldAttributes

  LoadValidFields(p, play_fields, data)

  data_pre_tasks, contains_pre_tasks := data["pre_tasks"]
  if contains_pre_tasks {
    td, _ := data_pre_tasks.([]interface{})
    p.Pre_tasks = LoadListOfBlocks(td, p, p, false)
  }
  data_tasks, contains_tasks := data["tasks"]
  if contains_tasks {
    td, _ := data_tasks.([]interface{})
    p.Tasks = LoadListOfBlocks(td, p, p, false)
  }
  data_post_tasks, contains_post_tasks := data["post_tasks"]
  if contains_post_tasks {
    td, _ := data_post_tasks.([]interface{})
    p.Post_tasks = LoadListOfBlocks(td, p, p, false)
  }
}

func (p *Play) Compile() []Block {
  flush_meta_data := map[interface{}]interface{} {
    "meta": "flush_handlers",
  }
  flush_block := NewBlock(flush_meta_data, p, nil, false)

  block_list := make([]Block, 0)
  block_list = append(block_list, p.Pre_tasks...)
  block_list = append(block_list, *flush_block)
  // FIXME: roles
  block_list = append(block_list, p.Tasks...)
  block_list = append(block_list, *flush_block)
  block_list = append(block_list, p.Post_tasks...)
  block_list = append(block_list, *flush_block)
  return block_list
}

func (p *Play) EvaluateTags(only_tags []string, skip_tags []string) bool {
  return EvaluateTags(p, only_tags, skip_tags)
}

// local getters
func (p *Play) Hosts() []string {
  if res, ok := p.Attr_hosts.([]string); ok {
    return res
  } else {
    res, _ := play_fields["hosts"].Default.([]string)
    return res
  }
}
func (p *Play) GatherFacts() bool {
  if res, ok := p.Attr_gather_facts.(bool); ok {
    return res
  } else {
    res, _ := play_fields["gather_facts"].Default.(bool)
    return res
  }
}
func (p *Play) GatherSubset() []string {
  if res, ok := p.Attr_gather_subset.([]string); ok {
    return res
  } else {
    res, _ := play_fields["gather_subset"].Default.([]string)
    return res
  }
}
func (p *Play) Serial() []int {
  if res, ok := p.Attr_serial.([]int); ok {
    return res
  } else {
    res, _ := play_fields["serial"].Default.([]int)
    return res
  }
}

func NewPlay(data map[interface{}]interface{}) *Play {
  p := new(Play)
  p.Load(data)
  p.RemovedHosts = make(map[string]bool)
  return p
}

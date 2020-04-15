package dao

import (
	"bytes"
	"fmt"
	"strings"
)

func (s *Session) Query(query string, args ...interface{}) *Session {
	s.querySrcData = &sqlData{Query: query, Args: args}
	return s
}

func (s *Session) Select(field ...string) *Session {
	return s.Cols(field...)
}

func (s *Session) addPart(p expPart) {
	s.parts = append(s.parts, p)
}

func (s *Session) Limit(v ...int64) *Session {
	x := make([]string, len(v))
	for k, a := range v {
		x[k] = fmt.Sprintf("%v", a)
	}
	s.addPart(ExpPart("limit", x...))
	return s
}

func (s *Session) OrderBy(v ...string) *Session {
	s.addPart(ExpPart("order by", v...))
	return s
}

func (s *Session) GroupBy(v ...string) *Session {
	s.addPart(ExpPart("group by", v...))
	return s
}

func (s *Session) Count() (int64, error) {
	s.Select("count(1) as cnt")
	x := make(map[string]int64)
	if err := s.queryCtx().Get(&x); err != nil {
		return 0, err
	}
	return x["cnt"], nil
}

func (s *Session) queryCtx() *QueryContext {
	sqls, sqlv := s.buildQuery()
	s.logOutput(sqls, sqlv)
	if s.tx != nil {
		rows, err := s.tx.Query(sqls, sqlv...)
		return &QueryContext{lastErr: err, rows: rows}
	}
	rows, err := s.dao.DB().Query(sqls, sqlv...)
	return &QueryContext{lastErr: err, rows: rows}
}

func (s *Session) Find(obj interface{}) error {
	return s.queryCtx().Find(obj)
}

func (s *Session) Get(obj interface{}) error {
	s.Limit(1)
	return s.queryCtx().Get(obj)
}

func (s *Session) buildQuery() (string, []interface{}) {
	if s.querySrcData != nil {
		return s.querySrcData.Query, s.querySrcData.Args
	}
	if s.table == "" {
		panic("tablename faild")
	}
	cols := "*"
	if len(s.fields) != 0 {
		cols = strings.Join(s.fields, ", ")
	}
	str := bytes.NewBuffer(nil)
	str.WriteString(fmt.Sprintf("select %s from %s", cols, s.table))
	condstr, condargs := s.cond.Build()
	str.WriteString(condstr)
	for _, v := range s.parts {
		str.WriteString(v.String())
	}
	return str.String(), condargs
}
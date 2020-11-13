package main

import "database/sql"

func rows2maps(rows *sql.Rows) (res []map[string]interface{}) {

	defer rows.Close()
	cols, _ := rows.Columns()
	cache := make([]interface{}, len(cols))
	// 为每一列初始化一个指针
	for index, _ := range cache {
		var a interface{}
		cache[index] = &a
	}

	for rows.Next() {
		rows.Scan(cache...)
		row := make(map[string]interface{})
		for i, val := range cache {

			// 处理数据类型
			v := *val.(*interface{})
			switch v.(type) {
			case []uint8:
				v = string(v.([]uint8))
			case nil:
				v = ""
			}
			row[cols[i]] = v
		}

		res = append(res, row)
	}

	return res
}

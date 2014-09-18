using System;
using System.Collections.Generic;

namespace gauge_csharp_lib
{
    public class Table
    {
        private readonly List<String> _headers;
        private readonly List<List<String>> _rows;

        public Table(List<String> headers)
        {
            this._headers = headers;
            _rows = new List<List<String>>();
        }

        public void AddRow(List<String> row)
        {
            _rows.Add(row);
        }

        public List<String> GetColumnNames()
        {
            return _headers;
        }

        public List<List<String>> GetRows()
        {
            return _rows;
        }
    }
}
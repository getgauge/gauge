using System;
using System.Collections.Generic;

namespace gauge_csharp_lib
{
    public class StepImpl
    {
        [Step("say <greeting> to <someone>")]
        public void sayHello(String greeting, String name)
        {
            Console.Out.WriteLine("{0}, {1}!!!!!!", greeting, name);
        }

        [Step("step with <table>")]
        public void newStep(Table table)
        {
            Console.Out.WriteLine("inside table step");
            List<string> columnNames = table.GetColumnNames();
            columnNames.ForEach(s => Console.Out.Write(s));
            Console.Out.WriteLine("");
            table.GetRows().ForEach(row => row.ForEach(cell => Console.Out.Write(cell)));
            
        }
    }
}
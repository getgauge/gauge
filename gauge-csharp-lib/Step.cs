using System;
using System.Collections.Generic;
using System.Linq;
using System.Text;
using System.Threading.Tasks;

namespace gauge_csharp_lib
{
    [AttributeUsage(AttributeTargets.Method)]
    public class Step : System.Attribute
    {
        public readonly string stepText ;

        public Step(string stepText)
        {
            this.stepText = stepText;
        }

        public string Name
        {
            get
            {
                return stepText;
            }
        }
    }
}

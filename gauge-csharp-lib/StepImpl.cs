using System;

namespace gauge_csharp_lib
{
    public class StepImpl
    {
        [Step("say <greeting> to <someone>")]
        public void sayHello()
        {
            Console.Out.WriteLine("hello world");
        }

        [Step("foo")]
        public void newStep()
        {
            Console.Out.WriteLine("Another step");
        }
    }
}
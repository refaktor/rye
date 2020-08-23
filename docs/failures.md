# Failures and Errors

When I started writing Rye I had no specific ideas about Exception handling. But the classic _try catch_ model looked annoying to me.

I had a problem with it visually, the code without exceptions flowed, but if you added all exception handling that needed to be there (sometimes) the 
code became a mess. Try catch added code structure, where there logic wise wasn't one. And many times (or in many languages) exception handling looks to me
like a goto statement.

I wanted to make something that is visually and structurally not so obtrusive. I wanted something in-flow. I composed something based on that, we have yet to
see if it really works any better than the classic approach.

Everything below this point is just a current hypothesis, it may be wrong in parts or alltogether, but let's play with it and see.

## A fresh look

Let's try to start from zero. 

  * Some exceptions are a result of __programming bug__. It means the program should stop (as we don't know what will happen next and what should). Exception should
  becommunicated to the user and logged for programmer, to fix the bug. We don't catch and handle these exceptions, we fix them if we know for them.
    * static bugs
      * syntax errors  - the code is not loadable rye, example: "err: {123 asda }"
      * naming errors - words that aren't defined or are misnamed
      * structure errors - the code isn't structured as used words would need it to be: "loop 2 3" ... 3 should be a block of code
    * runtime bugs

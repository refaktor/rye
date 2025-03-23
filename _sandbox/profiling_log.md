
# 20250322

I tested loop 10million in python, ruby, awk and all 3 take aroung 450ms on my computer (when on charger).

Yesterday Rye0 took around 800ms. Then I implemented builtins stapling into code blocks, optimised loop and add (small efect), avoided allocation of result, in Rye0 and Rye and now loop 10mil in also takes around 450ms! Rye takes around 800ms. 

Part of it can be the callbuiltin with currying ... try removing that to see ... part the evaluator itself with op/pipe words ..

Next I should clean the evaluator, make sense of it anyways and test where exactly we loose those 350ms. 

To benchmark 

* Rye code I used: test_adnums.rye , 
* to test pure Go constructed block where I already got 450ms I tested cmd/loop_benchark.go code that uses Rye Rye0 and FPVM versions

Could stamped in blocks cause any unexpected behaviour ... we will have to think about the edge cases ... will it be implicit... always there
will other stamping also happen ... const values for example (will we declare them or infer them?) ... context will be able to be locked and so could words be inside a specific context probably ... with same flah and same wordage x: final 10

This will make things faster, but also add certainty ... so that even if you use a mod word, it won't work: final fn { x }

Still not sure if we can just find a certain way to compile Rye code (op/pipe words, curry to) Rye0 when stamping values and always evaluate that. That would make all code faster and evaluator simpler.

Another speedup would be, at least in smaller simple functions, if we replaced contexts as variable lookup with static slots.

Like the trick with addnums where we dont create new integer, becuse it's pased by value and already recreated. Or loop index where we update existing integer we pass in (check for side effects) ... they don't have cruicial effect but avoiding of env.Integer allocation was noticable 5-10%

Another method would be to fist solve the most common paths, like we did at callbuiltin so they work the fastest.

* because all Rye does is call builtins in very case ... if we make things like loop + if as fast as python/ruby ... will at things that are not part of the language itself py/rby but ordinary functions (not so statically defined), or even FFI Rye actually be faster than those? 
# TODO

This a list of ideas to improve this application.

## v1

### Improve UX of adding new events to schedule

Instead of only adding to end, we should consider support adding them inline and
then shifting the following events down. There's not a good UX that comes to
mind though.

### Improve UX of aggregates of core/push/pull

It's not pretty to just have an unordered list. Think about this.

### Page for lift entry

### Toilsome in auto-incrementing lifts for each cycle

### Page for inputing routines

## v2

### Overview

Big problem is that fundamentally, there's too much effort. I want to show up to my workout, visit this application (for the first time!), click one button to start the workout, and start working out. All of this should effectively be less than 10 seconds. Here's what I think this means:

* Get rid of pre-defined schedule. This takes time and is tedious.
* Have a default routine that is just any lift (full-body workout effectively). Having to select or define a routine takes time.
* Automatically enter lift selection phase when starting a workout, which, if following a routine, recommend lifts according to the routine with weights defined by that routine if possible.
* If not, recommend lifts based on the following:
  1. If the lifts already been done this workout (exclude!).
  1. Muscle group that yet to be done.
  1. Lifts done most recently.
  1. Whether it's a more "common" lift. Like barbell squat versus hack squat.

Then, we enter set entry phase, which should allow inputing reps+weight quickly or workout conclusion. Consider setting a timer as well. Once done with sets, there should be a button to conclude the lift, at which point you're redirected to the lift entry phase again. A button to conclude workout should be omnipresent along with a timer.

### Additional Considerations

* If I start a workout and forget to stop it, it should automatically stop for me based on some ttl (3 hours?).
  * If I'm keeping track of workout time, how long do I count this workout as? The time of the last set entry? 3 hours? Don't count the workout time?
* Consider multi-tenancy. What happens if someone else wants to use the same app? Also, it means better support as a managed solution instead of needing to self-host.
  * Tables should support a notion of uid for user-level data.
* Support supersetting. That is, support entering sets for multiple lifts simultaneously.

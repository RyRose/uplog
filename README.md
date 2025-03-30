# uplog

## Database Schema

### Progress

- ID (auto-increment PK)
- date
- lift
- x2
- x2+45
- x2+55
- x2+60
- x1
- :100p:
- :poop:
- number of sets
- number of reps

### Schedule

- Date (PK)
- Workout
- Notes

### Schedules

- name (PK)
- day index (PK)
- workout

### Routines

- Name (PK)
- Steps
- Lift

### RoutineWorkoutMapping

- Routine (PK)
- Workout (PK)

### Workouts

- Name (PK)
- Template

### Subworkouts

Intended for denoting lifts from one workout are part of another.

- Subworkout (PK)
- Superworkout (PK)

### LiftWorkoutMapping

- Lift (PK)
- Workout (PK)

### Lifts

- Name (PK)
- Link
- Weight Type
- Notes

### LiftMuscleMapping

- Lift (PK)
- Muscle (PK)
- Movement

### Muscles

- Name (PK)
- Link
- Message (MRV/MEV/...)

### Muscle Movements

- Name (PK)
- Alias

### Side Weights

- Name (PK)
- Multiplier
- Addend
- Format

### Custom Template Variables

- Name
- Value

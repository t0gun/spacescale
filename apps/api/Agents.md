This file guides agents on how to work on this code base when needed for boiler plat tasks.
There are few things agents needs to know before producing output.

- we prioritize maintainability rather than just fix or just work.
- This default is blackbox testing and won't test private packages for simplicity
- we use a build file called make for builds, tests and local development
- our dev environment is goland so things like govet, ci lint and fmt are handled at IDE level rather than in CI
- In everything we do, we find way to write less code while still prioritizing simplicity and clarity
- For our tests, we don't enforce style, we chose the right style based on what we want to test.
- we use table driven tests where appropriate, subtests where its needed and normal tests where useful

#### To Agents:

We won't restrict you to not limit your creativity but our values should be put into consideration so we can
learn from each other 
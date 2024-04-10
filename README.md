# ottr
An Open Train Tracker, Reimagined.

ottr _(pronounced "otter")_ is a simple, small, and standalone program written in Go that periodically sends requests to the New York City Transit API (MTA) and outputs the information in a much more legible, ASCII-style map.

---

It was made as a school project in early 2023 with ambitious goals:
- Modular map support (sort of finished?)
- Support for any GTFS API
- A pretty GUI

However, due to time constraints and other responsabilities, ottr slipped away as a focus. It has been uploaded to GitHub as a means to showcase for my resume and portfolio. While there is currently no plans to complete this project, it does _technically_ work at the moment, and there may be a point in time where I change my mind and start working on it again - who knows!

### Building
To compile and run this program, follow the steps below:

1. Create an account with MTA to get your NYCT API key, and place it in `ottr.go` in place of the `PUT YOUR API KEY HERE` field.
2. Install the respective go packages for your system.
3. Run `go mod tidy` to install prerequisites, followed by `go build` to build it.
4. Success! A binary should be placed alongside the project files that you can run. Make sure you have the `map.csv` file placed in the same folder as the binary, or the program will not run.
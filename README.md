# Running the scraper
Build the scraper and API docker images, from the root of the project:
```bash
docker build -f docker/api/Dockerfile -t api
docker build -f docker/scraper/Dockerfile -t commoncrawl .
```

Run the API:
```bash
docker run --network=host api
```

Run the scraper:
```bash
docker run -e AWS_ACCESS_KEY_ID=<YOUR_ACCESS_KEY_HERE> -e AWS_SECRET_ACCESS_KEY=<YOUR_SECRET_HERE> \
-v /mnt/x/Test Data/Insight/BigDataTest:/app/data commoncrawl
```
Note that you'll need to replace <YOUR_ACCESS_KEY_HERE> and <YOUR_SECRET_HERE> with your AWS access key ID and secret access key, respectively. Additionally, you'll need to replace /mnt/x/Test Data/Insight/BigDataTest with the path to the directory where you want to save the scraped data.

# Specify the requested MB of each file type
You can specify the MB limit of each file type by editing the <b>parser/parser.go</b> file. In this file, you'll see a mimeLimits map that looks like this:
```go
var mimeLimits = map[string]int64{
	"image/jpeg":      20000,
	"image/jpg":       0, // It seems nothing is being detected as jpg, use jpeg instead
	"image/png":       20000,
	"application/pdf": 20000,
	"video/mp4":       20000,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   10000,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         5000,
	"application/vnd.openxmlformats-officedocument.presentationml.presentation": 5000,
}
```
To change the limit of a file type, simply update the corresponding value in this map. The value should be specified in MB.

# Number of threads
You can change the number of threads that the scraper should use by editing the Dockerfile for the scraper container. In this file, you'll see a CMD instruction that looks like this:
```Dockerfile
CMD ["go", "run", "main.go", "1000"]
```
The last argument ("1000") specifies the number of threads that the container should use. You can change this value to any number that you like. Note that this will impact the number of open sockets and concurrently open files, so choose a value that works well for your system.

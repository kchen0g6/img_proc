require "net/http"
require "uri"

class GoClient
  #sends data over HTTP to Go server on port 8080 
  #blocks until go finishes and returns job id 
  BASE = ENV.fetch("GO_INGEST_URL", "http://localhost:8080")

  def self.process(tempfile, filename, operation)
    uri = URI("#{BASE}/jobs")

    form = [
      ["archive", tempfile, { filename: filename, content_type: "application/zip" }],
      ["operation", operation]
    ]

    req = Net::HTTP::Post.new(uri)
    req.set_form(form, "multipart/form-data")

    res = Net::HTTP.start(uri.host, uri.port, read_timeout: 3600) do |http|
      http.request(req)
    end

    raise "go_proc returned #{res.code}" unless res.is_a?(Net::HTTPSuccess)

    JSON.parse(res.body).fetch("job_id")
  end
end
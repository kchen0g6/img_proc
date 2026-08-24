class JobsController < ApplicationController
  OPERATIONS = %w[ocr compress].freeze

  def create
    '''
    fields in our form arrive. 
    param is a rails provided object holding whatever cam with the request
    '''
    archive   = params[:archive]
    operation = params[:operation]

    if archive.blank?
      flash[:alert] = "Pick a zip file."
      return redirect_to root_path
    end

    unless OPERATIONS.include?(operation)
      flash[:alert] = "Pick an operation."
      return redirect_to root_path
    end

    job_id = GoClient.process(archive.tempfile, archive.original_filename, operation)

    # Ends request and tells the brwoser to go to (get "jobs/:id) /jobs/jobid
    # redirects to results page with the download and delete button
    redirect_to job_path(job_id)

  rescue => e
    flash[:alert] = "Processing failed: #{e.message}"
    redirect_to root_path
  end

  def show
    @job_id = params[:id]
  end

  def download
    path = results_path_for(params[:id])
    return head :not_found unless File.exist?(path)

    send_file path, filename: "#{params[:id]}.json", type: "application/json"
  end

  def destroy
    path = results_path_for(params[:id])
    File.delete(path) if File.exist?(path)

    redirect_to root_path, notice: "Deleted."
  end

  private

  def results_path_for(id)
    Rails.root.join("..", "results", "#{File.basename(id)}.json").to_s
  end
end
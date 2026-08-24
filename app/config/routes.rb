Rails.application.routes.draw do
  #get "home/index"
  '''
  At the default browser app it goes to home controller and runs method 
  specified after # to be run. 
  
  From "home#index" it builds the path views/home/index.html.erb and 
  renders the web page,

  '''
  root "home#index"

  post "jobs" => "jobs#create"
  get "jobs/:id" => "jobs#show", as: :job
  get "jobs/:id/download" => "jobs#download", as: :download_job
  delete "jobs/:id" => "jobs#destroy"

  '''
  Generates routes
  Prefix        Verb    URI Pattern                    Controller#Action
  root          GET     /                              home#index

  jobs          POST    /jobs(.:format)                jobs#create
  job           GET     /jobs/:id(.:format)            jobs#show
  download_job  GET     /jobs/:id/download(.:format)   jobs#download
                DELETE  /jobs/:id(.:format)            jobs#destroy
  '''

  
  # Define your application routes per the DSL in https://guides.rubyonrails.org/routing.html

  # Reveal health status on /up that returns 200 if the app boots with no exceptions, otherwise 500.
  # Can be used by load balancers and uptime monitors to verify that the app is live.
  get "up" => "rails/health#show", as: :rails_health_check

  # Render dynamic PWA files from app/views/pwa/* (remember to link manifest in application.html.erb)
  # get "manifest" => "rails/pwa#manifest", as: :pwa_manifest
  # get "service-worker" => "rails/pwa#service_worker", as: :pwa_service_worker

  # Defines the root path route ("/")
  # root "posts#index"
end

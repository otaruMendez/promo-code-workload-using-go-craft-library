- Database Tables
    - promotions
        - pk (pk, primary key used to identify a record)
        - id (id from the csv)
        - price (price from the csv)
        - expiration_date (expiration date from the csv)
        - version (version is linked to the scheduler reponsible for inserting this record)

- promo_scheduler recurring job (runs every 30mins)
    - each run has a schedulerID used to uniquely identify a schdule job
    - calculates the number of lines from the promotions csv file
    - creates insert_promo jobs to insert range of records into the promotions table

- insert_promo job workers
    - for each job, get the range of records from the promotions csv file
    - upsert these records in the promotions table including the scheduler id

- GET endpoint to fetch a promotion via the pk

- How would you operate this app in production (e.g. deployment, scaling, monitoring)?
    - Scaling: increasing the worker pool (so it can run more jobs concurrently)
    - Monitoring: check the state of the jobs

- How to Run the project
    - Install docker
    - Run `docker compose up -d`
    - Access the API via localhost:8080 e.g to get a promo info - do a GET http://localhost:8080/promotions/1

- Monitoring the Jobs Items
    - Follow instructions here: https://github.com/gocraft/work#run-the-web-ui

Technologies:
    - Redis (for background jobs using this library https://github.com/gocraft/work)
    - PostgresQL (for promotions data)

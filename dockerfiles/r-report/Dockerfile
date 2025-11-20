FROM rocker/tidyverse:latest

# Install additional R packages needed for the project
RUN R -e "install.packages(c('sendmailR', 'RPostgreSQL', 'DBI', 'dplyr', 'knitr', 'rmarkdown', 'ggplot2'), repos='https://cran.rstudio.com/')"

# Create app directory
WORKDIR /app

# Copy R script
COPY r-scripts/report_generation.R /app/

# Install additional system dependencies if needed
RUN apt-get update && apt-get install -y \
    texlive-latex-base \
    texlive-fonts-recommended \
    texlive-latex-extra \
    lmodern \
    && rm -rf /var/lib/apt/lists/*

# Run the R script when the container starts
CMD ["Rscript", "report_generation.R"]
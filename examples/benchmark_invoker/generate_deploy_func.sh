#!/bin/bash

yaml_template="benchmark.yaml"
yaml_dir="func_yamls"
endpoints_file="benchmark_endpoints.txt"

if [ $# -ne 1 ]; then
  echo "Usage: $0 <num_functions>"
  exit 1
fi

num_func=$1

mkdir -p ${yaml_dir}
rm ${yaml_dir}/* # Clear if not empty
rm ${endpoints_file} # Remove endpoints file

# Generate the yaml for ${num_func} functions and deploy them to vHive
for (( i=0; i<${num_func}; i++ )); do
  next_function_name="benchmark-func-${i}"
  next_function_file="${yaml_dir}/${next_function_name}.yaml"
  sed "s/FUNCTION_NAME/${next_function_name}/" "${yaml_template}" > ${next_function_file}
done

echo "Done generating function yamls"

# First delete all functions that have been deployed to vHive
kn service delete --all
sleep 3

# Now deploy all functions to vHive
for f in ${yaml_dir}/*.yaml; do
  bname=`basename ${f}`
  name="${bname%.*}"
  echo "Deploying ${name}"

  kubectl apply --filename ${f}
  sleep 1

  # Get endpoint
  raw_endpoint=$(kn service describe ${name} -o url)

  # Strip off http part, add port (:80), and write to file
  cleaned_endpoint=`echo ${raw_endpoint} | sed 's/https\?:\/\///'`
  echo "${cleaned_endpoint}:80" >> ${endpoints_file}
done

echo "Done deploying functions. Endpoints are in ${endpoints_file}"

#!/usr/bin/python3
from setuptools import setup, find_packages

setup(
    name='prometheus-ss-exporter',
    python_requires='>3.5.2',
    version='1.0.0',

    description='prometheus socket statistics exporter',

    url='https://github.com/cherusk/prometheus_ss_exporter',

    author='Matthias Tafelmeier',
    author_email='matthias.tafelmeier@gmx.net',

    license='MIT',

    classifiers=[
        'License :: OSI Approved :: MIT License',
        'Programming Language :: Python :: 3',
    ],

    keywords=[
        'monitoring',
        'exporter',
        'prometheus',
        'linux',
        'socket statistics',
        'kernel statistics',
        'network stack',
        'ss2'
    ],

    packages=find_packages(),

    install_requires=[
        'pyroute2==0.5.10',
        'prometheus-client',
        'PyYAML==5.1.1',
        'psutil'
    ],

    entry_points={
        'console_scripts': [
            'prometheus_ss_exporter=prometheus_ss_exporter.__init__:main',
        ],
    },
)
